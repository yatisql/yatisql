use crate::io::error::IoError;
use crate::io::file_reader::FileReader;
use crate::utils::col_name_to_index;

pub struct TabularFileReader<'row> {
    reader: FileReader,
    header: Vec<Vec<u8>>,
    fields: Vec<&'row [u8]>,
    delimiter: u8,
    has_headers: bool,
}

impl<'row> TabularFileReader<'row> {
    /// Open a CSV file, optionally reading headers.
    pub fn open(path: &str, has_headers: bool, delimiter: u8) -> Result<Self, IoError> {
        let reader = FileReader::open(path)?;
        let mut tab_reader = TabularFileReader {
            reader,
            delimiter,
            has_headers,
            header: Vec::new(),
            fields: Vec::new(),
        };
        // tab_reader.read_header()?;
        Ok(tab_reader)
    }

    // fn read_header(&mut self) -> Result<(), IoError> {
    //     if let Some(line) = self.reader.next_line()? {
    //         let mut start = 0;
    //         let mut i = 0;
    //         let mut count = 0;
    //         while i <= line.len() {
    //             if i == line.len() || line[i] == self.delimiter {
    //                 if self.has_headers {
    //                     self.header.push(line[start..i].to_vec());
    //                 }
    //                 start = i + 1;
    //                 count += 1;
    //             }
    //             i += 1;
    //         }
    //         self.fields.resize(count, b"")
    //     }
    //     if !self.has_headers {
    //         self.reader.rewind()?;
    //     }
    //     Ok(())
    // }

    pub fn col_name_to_index(&self, column_name: &[u8]) -> Option<usize> {
        for (i, hdr) in self.header.iter().enumerate() {
            if hdr.as_slice() == column_name {
                return Some(i);
            }
        }
        col_name_to_index(column_name)
    }

    // pub fn next_row(&mut self) -> Result<Option<&[&[u8]]>, IoError> {
    //     if let Some(line) = self.reader.next_line()? {
    //         self.fields.clear();
    //         let mut start = 0;
    //         let mut i = 0;
    //         while i <= line.len() {
    //             if i == line.len() || line[i] == self.delimiter {
    //                 self.fields.push(&line[start..i]);
    //                 start = i + 1;
    //             }
    //             i += 1;
    //         }
    //         Ok(Some(&self.fields))
    //     } else {
    //         Ok(None)
    //     }
    // }
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

    // #[test]
    // fn test_with_headers() {
    //     let content = "col1,col2\nval1,val2\nval3,val4\n";
    //     let file = create_temp_file(content);
    //     let path = file.path().to_str().unwrap();
    //     let mut reader = TabularFileReader::open(&path, true, b',').unwrap();
    //
    //     assert_eq!(reader.header[0], b"col1");
    //     assert_eq!(reader.header[1], b"col2");
    //
    //     let mut row = reader.next_row().unwrap().unwrap();
    //     assert_eq!(row[0], b"val1");
    //     assert_eq!(row[1], b"val2");
    //
    //     row = reader.next_row().unwrap().unwrap();
    //     assert_eq!(row[0], b"val3");
    //     assert_eq!(row[1], b"val4");
    //
    //     // assert!(reader.next_row().unwrap().is_none());
    //
    //     std::fs::remove_file(path).unwrap();
    // }

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
    //     assert!(reader.header.is_none() || reader.header.as_ref().unwrap().is_empty());
    //     assert!(reader.next_row().unwrap().is_none());
    //     std::fs::remove_file(path).unwrap();
    // }
}
