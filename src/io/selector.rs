use std::ops;
use crate::io::error::IoError;
use crate::io::traits::{LineReader, BUFFER_SIZE};
use crate::utils::{col_name_to_index, next_delimiter_pos};

pub struct Selector {
    reader: Box<dyn LineReader>,
    buffer: Vec<u8>,
    header: Vec<Vec<u8>>,
    rules: Vec<usize>,
    rules_to_slices_binding: Vec<ops::Range<usize>>,
    delimiter: u8,
}

impl LineReader for Selector {
    fn next_line(&mut self, buffer: &mut [u8]) -> Result<ops::Range<usize>, IoError> {
        let line_range = self.reader.next_line(&mut self.buffer)?;
        let line = &self.buffer[line_range];

        let mut prev_line_index: usize = 0;

        let mut column_index: usize = 0;
        loop {
            let line_index = next_delimiter_pos(line, prev_line_index, self.delimiter);

            for (rules_index, &col_index) in self.rules.iter().enumerate() {
                if col_index == column_index {
                    self.rules_to_slices_binding[rules_index] = prev_line_index..line_index;
                }
            }

            column_index += 1;
            prev_line_index = line_index + 1;
            if line_index >= line.len() {
                break;
            }
        }

        let mut out_buf_index = 0;
        for slice in self.rules_to_slices_binding.iter() {
            buffer[out_buf_index..out_buf_index + (slice.end - slice.start)]
                .copy_from_slice(&line[slice.clone()]);
            buffer[out_buf_index + (slice.end - slice.start)] = self.delimiter;
            out_buf_index = out_buf_index + (slice.end - slice.start) + 1;
        }
        Ok(0..out_buf_index-1)
    }
}

impl Selector {
    /// Open a CSV file, optionally reading headers.
    pub fn new(reader: Box<dyn LineReader>, delimiter: u8) -> Result<Self, IoError> {
        Ok(Selector {
            reader,
            buffer: vec![0u8; BUFFER_SIZE],
            header: Vec::new(),
            rules: Vec::new(),
            rules_to_slices_binding: Vec::new(),
            delimiter,
        })
    }

    pub fn read_header(&mut self) -> Result<(), IoError> {
        let line_range = self.reader.next_line(&mut self.buffer)?;
        let mut prev_i = 0;
        for i in line_range.clone() {
            if self.buffer[i] == self.delimiter {
                self.header.push(self.buffer[prev_i..i].to_vec());
                prev_i = i + 1;
            } else if i == line_range.end - 1 {
                self.header.push(self.buffer[prev_i..i+1].to_vec());
            }
        }
        Ok(())
    }

    pub fn set_rules(&mut self, columns: &[&[u8]]) -> Result<(), IoError>  {
        for column_name in columns {
            if let Some(idx) = self.col_name_to_index(column_name) {
                self.rules.push(idx);
            } else {
                return Err(IoError::ColumnNotFound(format!("{:?}", column_name)));
            }
        }
        self.rules_to_slices_binding.resize(self.rules.len(), 0..0);
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
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;
    use crate::io::file_reader::FileReader;

    fn setup(content: &[u8], selecting: &[&[u8]]) -> Selector {
        let mut file = NamedTempFile::new().expect("Failed to create temp file");
        file.write(content).expect("Failed to write to temp file");
        let path = file.path().to_str().unwrap();
        let reader = FileReader::open(path).unwrap();
        let mut selector = Selector::new(Box::new(reader), b',').unwrap();
        selector.read_header().unwrap();
        selector.set_rules(selecting).unwrap();
        selector
    }

    #[test]
    fn test_with_headers() {
        let content = b"col1,col2\nval1,val2\nval3,val4\n";
        let mut buffer = vec![0u8; 20];
        let mut selector = setup(content, &[b"col2", b"col1"]);

        for result in [b"val2,val1", b"val4,val3"] {
            let range = selector.next_line(&mut buffer).unwrap();
            assert_eq!(&buffer[range], result);
        }
        assert!(matches!(selector.next_line(&mut buffer), Err(IoError::EndOfFile)));
    }

    #[test]
    fn test_col_name_to_index() {
        let content = b"foo,bar,baz\n1,2,3\n";
        let selector = setup(content, &[b"baz", b"foo"]);
        assert_eq!(selector.col_name_to_index(b"foo"), Some(0));
        assert_eq!(selector.col_name_to_index(b"bar"), Some(1));
        assert_eq!(selector.col_name_to_index(b"baz"), Some(2));
        assert_eq!(selector.col_name_to_index(b"qux"), None);

        assert_eq!(selector.col_name_to_index(b"A"), Some(0));
        assert_eq!(selector.col_name_to_index(b"B"), Some(1));
        assert_eq!(selector.col_name_to_index(b"C"), Some(2));
        // 'D' does not exist in header, needs an additional check
        assert_eq!(selector.col_name_to_index(b"D"), Some(3));
    }
}
