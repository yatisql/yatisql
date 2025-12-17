use std::fs::File;
use std::io::{Read, Seek};
use crate::io::error::IoError;

const BUFFER_SIZE: usize = 1024 * 1024 * 1; // 1 MB buffer

pub struct BufferedFileReader {
    file: File,
    buffer: Vec<u8>,
    buf_start: usize,
    buf_end: usize,
    eof: bool,
}

impl BufferedFileReader {
    pub fn open(path: &str) -> Result<Self, IoError> {
        let file = File::open(&path).map_err(IoError::from)?;
        Ok(BufferedFileReader {
            file,
            buffer: vec![0; BUFFER_SIZE],
            buf_start: 0,
            buf_end: 0,
            eof: false,
        })
    }

    /// Reads the next line from the file. Returns None if EOF is reached.
    /// Returns a slice into the internal buffer; valid until next read_line call.
    pub fn read_line(&mut self) -> Result<Option<&[u8]>, IoError> {
        loop {
            // If buffer is empty, fill it
            if self.buf_start == self.buf_end && !self.eof {
                let n = self.file.read(&mut self.buffer).map_err(IoError::from)?;
                if n == 0 {
                    self.eof = true;
                    return Ok(None);
                }
                self.buf_start = 0;
                self.buf_end = n;
            }
            // Search for newline in buffer
            if self.buf_start < self.buf_end {
                let buf = &self.buffer[self.buf_start..self.buf_end];
                if let Some(pos) = buf.iter().position(|&b| b == b'\n') {
                    let line_end = self.buf_start + pos;
                    let line = &self.buffer[self.buf_start..line_end];
                    self.buf_start = line_end + 1; // skip the newline
                    return Ok(Some(line));
                } else {
                    // No newline found, check if buffer is full
                    if self.buf_end - self.buf_start == BUFFER_SIZE {
                        // Line too long for buffer
                        return Err(IoError::LineTooLong);
                    } else if self.eof {
                        // Return the rest as the last line
                        let line = &self.buffer[self.buf_start..self.buf_end];
                        self.buf_start = self.buf_end;
                        return if line.is_empty() { Ok(None) } else { Ok(Some(line)) };
                    } else {
                        // Move remaining data to start and fill buffer
                        let remaining = self.buf_end - self.buf_start;
                        if remaining > 0 {
                            self.buffer.copy_within(self.buf_start..self.buf_end, 0);
                        }
                        self.buf_start = 0;
                        self.buf_end = remaining;
                        let n = self.file.read(&mut self.buffer[remaining..]).map_err(IoError::from)?;
                        if n == 0 {
                            self.eof = true;
                        }
                        self.buf_end += n;
                    }
                }
            } else if self.eof {
                return Ok(None);
            }
        }
    }

    pub fn rewind(&mut self) -> Result<(), IoError> {
        self.file.seek(std::io::SeekFrom::Start(0)).map_err(IoError::from)?;
        self.buf_start = 0;
        self.buf_end = 0;
        self.eof = false;
        Ok(())
    }

    pub fn is_eof(&self) -> bool {
        self.eof && self.buf_start == self.buf_end
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;

    fn write_temp_file(contents: &str) -> NamedTempFile {
        let mut file = NamedTempFile::new().expect("Failed to create temp file");
        write!(file, "{}", contents).expect("Failed to write to temp file");
        file
    }

    #[test]
    fn test_read_lines_basic() {
        let file = write_temp_file("line1\nline2\nline3\n");
        let path = file.path().to_str().unwrap();
        let mut reader = BufferedFileReader::open(path).unwrap();
        assert_eq!(reader.read_line().unwrap(), Some("line1".as_bytes()));
        assert_eq!(reader.read_line().unwrap(), Some("line2".as_bytes()));
        assert_eq!(reader.read_line().unwrap(), Some("line3".as_bytes()));
        assert_eq!(reader.read_line().unwrap(), None);
        assert!(reader.is_eof());
    }

    #[test]
    fn test_read_lines_no_trailing_newline() {
        let file = write_temp_file("foo\nbar\nbaz");
        let path = file.path().to_str().unwrap();
        let mut reader = BufferedFileReader::open(path).unwrap();
        assert_eq!(reader.read_line().unwrap(), Some("foo".as_bytes()));
        assert_eq!(reader.read_line().unwrap(), Some("bar".as_bytes()));
        assert_eq!(reader.read_line().unwrap(), Some("baz".as_bytes()));
        assert_eq!(reader.read_line().unwrap(), None);
        assert!(reader.is_eof());
    }

    #[test]
    fn test_empty_file() {
        let file = write_temp_file("");
        let path = file.path().to_str().unwrap();
        let mut reader = BufferedFileReader::open(path).unwrap();
        assert_eq!(reader.read_line().unwrap(), None);
        assert!(reader.is_eof());
    }

    #[test]
    fn test_single_line() {
        let file = write_temp_file("onlyline\n");
        let path = file.path().to_str().unwrap();
        let mut reader = BufferedFileReader::open(path).unwrap();
        assert_eq!(reader.read_line().unwrap(), Some("onlyline".as_bytes()));
        assert_eq!(reader.read_line().unwrap(), None);
        assert!(reader.is_eof());
    }

    #[test]
    fn test_long_line() {
        let long_line = "a".repeat(10000) + "\n";
        let file = write_temp_file(&long_line);
        let path = file.path().to_str().unwrap();
        let mut reader = BufferedFileReader::open(path).unwrap();
        assert_eq!(reader.read_line().unwrap(), Some("a".repeat(10000).as_bytes()));
        assert_eq!(reader.read_line().unwrap(), None);
        assert!(reader.is_eof());
    }
}
