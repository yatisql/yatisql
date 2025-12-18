use std::fs::File;
use std::io::{Seek};
use std::ops;
use std::os::unix::fs::FileExt;
use crate::io::error::IoError;
use crate::io::traits::LineReader;

pub struct FileReader {
    file: File,
    file_offset: u64,
    buf_start: usize,
    buf_end: usize,
}

impl LineReader for FileReader {
    fn next_line(&mut self, buffer: &mut [u8]) -> Result<ops::Range<usize>, IoError> {
        for i in self.buf_start..self.buf_end {
            if buffer[i] == b'\n' {
                let result = self.buf_start..i;
                self.buf_start = i + 1; // skip the newline
                return Ok(result);
            }
        }
        match self.load_buffer(buffer) {
            Err(IoError::EndOfFile) => {
                if self.buf_start < self.buf_end {
                    let result = self.buf_start..self.buf_end;
                    self.buf_start = self.buf_end;
                    Ok(result)
                } else {
                    Err(IoError::EndOfFile)
                }
            }
            Err(e) => Err(e),
            Ok(()) => Ok(self.next_line(buffer)?)
        }
    }
}

impl FileReader {
    pub fn open(path: &str) -> Result<Self, IoError> {
        let file = File::open(&path).map_err(IoError::from)?;
        Ok(FileReader {
            file,
            file_offset: 0,
            buf_start: 0,
            buf_end: 0,
        })
    }

    fn load_buffer(&mut self, buffer: &mut [u8]) -> Result<(), IoError> {
        buffer.copy_within(self.buf_start..self.buf_end, 0);
        self.buf_end = self.buf_end - self.buf_start;
        self.buf_start = 0;

        let available_buffer = &mut buffer[self.buf_end..];

        if available_buffer.is_empty() {
            return Err(IoError::LineTooLong);
        }

        let bytes_read = self.file.read_at(available_buffer, self.file_offset)?;
        self.file_offset += bytes_read as u64;
        self.buf_end += bytes_read;

        if bytes_read == 0 {
            return Err(IoError::EndOfFile);
        }
        Ok(())
    }

    #[allow(dead_code)]
    pub fn rewind(&mut self) -> Result<(), IoError> {
        self.file.seek(std::io::SeekFrom::Start(0)).map_err(IoError::from)?;
        self.buf_start = 0;
        self.buf_end = 0;
        self.file_offset = 0;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;

    fn setup(content: &[u8]) -> FileReader {
        let mut file = NamedTempFile::new().expect("Failed to create temp file");
        file.write(content).expect("Failed to write to temp file");
        let path = file.path().to_str().unwrap();
        let reader = FileReader::open(path).unwrap();
        reader
    }

    #[test]
    fn test_read_lines_basic() {
        let content = b"line1\nline2\nline3\n";
        let mut reader = setup(content);
        let mut buffer = vec![0u8; 6];

        for result in [b"line1", b"line2", b"line3"] {
            let range = reader.next_line(&mut buffer).unwrap();
            assert_eq!(&buffer[range], result);
        }
        assert!(matches!(reader.next_line(&mut buffer), Err(IoError::EndOfFile)));
    }

    #[test]
    fn test_read_lines_no_trailing_newline() {
        let content = b"foo\nbar\nbaz";
        let mut reader = setup(content);
        let mut buffer = vec![0u8; 4];

        for result in [b"foo", b"bar", b"baz"] {
            let range = reader.next_line(&mut buffer).unwrap();
            assert_eq!(&buffer[range], result);
        }
        assert!(matches!(reader.next_line(&mut buffer), Err(IoError::EndOfFile)));
    }

    #[test]
    fn test_empty_file() {
        let content = b"";
        let mut reader = setup(content);
        let mut buffer = vec![0u8; 20];
        assert!(matches!(reader.next_line(&mut buffer), Err(IoError::EndOfFile)));
    }
}
