use std::fs::File;
use std::io::Write;
use std::path::Path;
use crate::io::error::IoError;
use crate::io::traits::{LineReader};

pub struct FileWriter {
    reader: Box<dyn LineReader>,
    file: File,
}

impl LineReader for FileWriter {
    fn next_line(&mut self, buffer: &mut [u8]) -> Result<std::ops::Range<usize>, IoError> {
        self.write_next_line(buffer)
    }
}

impl FileWriter {
    pub fn new<P: AsRef<Path>>(reader: Box<dyn LineReader>, file_path: P) -> Result<Self, IoError> {
        let file = File::create(file_path).map_err(|e| IoError::from(e))?;
        Ok(FileWriter {
            reader,
            file,
        })
    }

    pub fn write_next_line(&mut self, buffer: &mut [u8]) -> Result<std::ops::Range<usize>, IoError> {
        match self.reader.next_line(buffer) {
            Ok(range) => {
                let line_data = &buffer[range.clone()];
                self.file.write(line_data).map_err(|e| IoError::from(e))?;
                self.file.write(b"\n").map_err(|e| IoError::from(e))?;
                self.file.flush().map_err(|e| IoError::from(e))?;
                Ok(range)
            }
            Err(e) => Err(e),
        }
    }

    #[allow(dead_code)]
    pub fn write_all_lines(&mut self, buffer: &mut [u8]) -> Result<(), IoError> {
        loop {
            match self.write_next_line(buffer) {
                Ok(_range) => {}
                Err(IoError::EndOfFile) => break,
                Err(e) => return Err(e),
            }
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::io::file_reader::FileReader;
    use std::io::{Read, Write};
    use tempfile::NamedTempFile;
    use crate::io::traits::BUFFER_SIZE;

    fn setup(content: &[u8]) -> FileWriter {
        let mut file = NamedTempFile::new().expect("Failed to create temp file");
        file.write(content).expect("Failed to write to temp file");
        let path = file.path().to_str().unwrap();
        let reader = FileReader::open(path).unwrap();
        let writer = FileWriter::new(Box::new(reader), "output.txt").unwrap();
        writer
    }

    fn read_file_content(path: &str) -> Vec<u8> {
        let mut file = File::open(path).expect("Failed to open file");
        let mut content = Vec::new();
        file.read_to_end(&mut content).expect("Failed to read file");
        content
    }

    #[test]
    fn test_write_all_lines() {
        let content = b"hello\nworld\nthis\nis\na\ntest\n";
        let mut writer = setup(content);
        let mut buffer = vec![0u8; BUFFER_SIZE];
        writer.write_all_lines(&mut buffer).unwrap();
        let output_content = read_file_content("output.txt");
        assert_eq!(output_content, content);
    }

}
