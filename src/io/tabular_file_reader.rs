use crate::io::buferred_file_reader::BufferedFileReader;
use crate::io::error::IoError;
use crate::utils::col_name_to_index;

pub struct TabularFileReader {
    reader: BufferedFileReader,
    header: Vec<Vec<u8>>,
    line_size: usize,
    has_headers: bool,
    delimiter: u8,
}

impl TabularFileReader {
    /// Open a CSV file, optionally reading headers.
    pub fn open(path: &str, has_headers: bool, delimiter: u8) -> Result<Self, IoError> {
        let reader = BufferedFileReader::open(path)?;
        let mut tab_reader = TabularFileReader {
            reader,
            header: Vec::new(),
            line_size: 0,
            has_headers,
            delimiter,
        };
        tab_reader.read_header()?;
        Ok(tab_reader)
    }

    fn read_header(&mut self) -> Result<(), IoError> {
        if let Some(line) = self.reader.read_line()? {
            let mut start = 0;
            let mut i = 0;
            while i <= line.len() {
                if i == line.len() || line[i] == self.delimiter {
                    if self.has_headers {
                        self.header.push(line[start..i].to_vec());
                    }
                    start = i + 1;
                    self.line_size += 1;
                }
                i += 1;
            }
        }
        if !self.has_headers {
            self.reader.rewind()?;
        }
        Ok(())
    }

    pub fn col_name_to_index(&self, column_name: &[u8]) -> Option<usize> {
        for (i, hdr) in self.header.iter().enumerate() {
            if hdr.as_slice() == column_name {
                return Some(i);
            }
        }
        col_name_to_index(column_name)
    }

    pub fn next_row(&mut self, buff: &mut [&[u8]]) -> Result<bool, IoError> {
        if let Some(line) = self.reader.read_line()? {
            let mut start = 0;
            let mut i = 0;
            let mut buf_i = 0;
            while i <= line.len() && buf_i < buff.len() {
                if i == line.len() || line[i] == self.delimiter {
                    buff[buf_i] = &line[start..i];
                    start = i + 1;
                }
                i += 1;
                buf_i += 1;
            }
            Ok(true)
        } else {
            Ok(false)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;

    fn create_temp_file(contents: &str) -> NamedTempFile {
        let mut file = NamedTempFile::new().expect("Failed to create temp file");
        write!(file, "{}", contents).expect("Failed to write to temp file");
        file
    }

    #[test]
    fn test_with_headers() {
        let content = "col1,col2\nval1,val2\nval3,val4\n";
        let file = create_temp_file(content);
        let path = file.path().to_str().unwrap();
        let mut reader = TabularFileReader::open(&path, true, b',').unwrap();

        assert_eq!(reader.header[0], b"col1");
        assert_eq!(reader.header[1], b"col2");

        let mut buff: [&[u8]; 2] = [&[]; 2];

        reader.next_row(&mut buff).unwrap();
        assert_eq!(buff[0], b"val1");
        assert_eq!(buff[1], b"val2");

        reader.next_row(&mut buff).unwrap();
        assert_eq!(buff[0], b"val3");
        assert_eq!(buff[1], b"val4");

        // assert!(reader.next_row().unwrap().is_none());

        std::fs::remove_file(path).unwrap();
    }

    // #[test]
    // fn test_without_headers() {
    //     let content = "val1,val2\nval3,val4\n";
    //     let file = create_temp_file(content);
    //     let path = file.path().to_str().unwrap();
    //     let mut reader = TabularFileReader::open(&path, false, b',').unwrap();
    //     // Header should be None
    //     assert!(reader.header.is_none());
    //     // Read first row
    //     {
    //         let row = reader.next_row().unwrap().unwrap();
    //         assert_eq!(row[0], b"val1");
    //         assert_eq!(row[1], b"val2");
    //     }
    //     // Read second row
    //     {
    //         let row = reader.next_row().unwrap().unwrap();
    //         assert_eq!(row[0], b"val3");
    //         assert_eq!(row[1], b"val4");
    //     }
    //     // No more rows
    //     assert!(reader.next_row().unwrap().is_none());
    //     std::fs::remove_file(path).unwrap();
    // }

    #[test]
    fn test_col_name_to_index() {
        let content = "foo,bar,baz\n1,2,3\n";
        let file = create_temp_file(content);
        let path = file.path().to_str().unwrap();
        let reader = TabularFileReader::open(&path, true, b',').unwrap();
        assert_eq!(reader.col_name_to_index(b"foo"), Some(0));
        assert_eq!(reader.col_name_to_index(b"bar"), Some(1));
        assert_eq!(reader.col_name_to_index(b"baz"), Some(2));
        assert_eq!(reader.col_name_to_index(b"qux"), None); // fallback to col_name_to_index fn
        std::fs::remove_file(path).unwrap();
    }

    // #[test]
    // fn test_empty_file() {
    //     let content = "";
    //     let file = create_temp_file(content);
    //     let path = file.path().to_str().unwrap();
    //     let mut reader = TabularFileReader::open(&path, true, b',').unwrap();
    //     assert!(reader.header.is_empty());
    //     assert!(reader.next_row().unwrap().is_none());
    //     std::fs::remove_file(path).unwrap();
    // }
}
