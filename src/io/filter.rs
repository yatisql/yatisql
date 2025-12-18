use std::ops;
use crate::io::error::IoError;
use crate::io::traits::{LineReader, BUFFER_SIZE};
use crate::utils::{col_name_to_index, next_delimiter_pos};

#[derive(Debug, Clone, Copy, PartialEq)]
pub enum ComparisonOp {
    Eq,      // =
    Ne,      // !=
    Lt,      // <
    Le,      // <=
    Gt,      // >
    Ge,      // >=
}

#[derive(Debug, Clone)]
pub enum Condition {
    Compare {
        column: Vec<u8>,
        op: ComparisonOp,
        value: Vec<u8>,
    },
    And(Box<Condition>, Box<Condition>),
    Or(Box<Condition>, Box<Condition>),
}

pub struct Filter {
    reader: Box<dyn LineReader>,
    buffer: Vec<u8>,
    header: Vec<Vec<u8>>,
    condition: Option<Condition>,
    delimiter: u8,
}

impl LineReader for Filter {
    fn next_line(&mut self, buffer: &mut [u8]) -> Result<ops::Range<usize>, IoError> {
        loop {
            let line_range = self.reader.next_line(&mut self.buffer)?;
            let line = &self.buffer[line_range.clone()];

            // If no condition is set, pass through all lines
            if self.condition.is_none() {
                buffer[..line.len()].copy_from_slice(line);
                return Ok(0..line.len());
            }

            // Evaluate the condition
            if self.evaluate_condition(line)? {
                buffer[..line.len()].copy_from_slice(line);
                return Ok(0..line.len());
            }
            // If condition is false, continue to next line
        }
    }
}

impl Filter {
    /// Create a new Filter with a reader and delimiter
    pub fn new(reader: Box<dyn LineReader>, delimiter: u8) -> Result<Self, IoError> {
        Ok(Filter {
            reader,
            buffer: vec![0u8; BUFFER_SIZE],
            header: Vec::new(),
            condition: None,
            delimiter,
        })
    }

    /// Read the header line from the input
    pub fn read_header(&mut self) -> Result<(), IoError> {
        let line_range = self.reader.next_line(&mut self.buffer)?;
        let mut prev_i = 0;
        for i in line_range.clone() {
            if self.buffer[i] == self.delimiter {
                self.header.push(self.buffer[prev_i..i].to_vec());
                prev_i = i + 1;
            } else if i == line_range.end - 1 {
                self.header.push(self.buffer[prev_i..i + 1].to_vec());
            }
        }
        Ok(())
    }

    /// Set the filtering condition
    pub fn set_condition(&mut self, condition: Condition) {
        self.condition = Some(condition);
    }

    /// Convert column name to index
    pub fn col_name_to_index(&self, column_name: &[u8]) -> Option<usize> {
        for (i, hdr) in self.header.iter().enumerate() {
            if hdr.as_slice() == column_name {
                return Some(i);
            }
        }
        col_name_to_index(column_name)
    }

    /// Get the value of a column from a line
    fn get_column_value(&self, line: &[u8], column_name: &[u8]) -> Result<Vec<u8>, IoError> {
        let col_index = self.col_name_to_index(column_name)
            .ok_or_else(|| IoError::ColumnNotFound(format!("{:?}", column_name)))?;

        let mut prev_line_index: usize = 0;
        let mut column_index: usize = 0;

        loop {
            let line_index = next_delimiter_pos(line, prev_line_index, self.delimiter);

            if col_index == column_index {
                return Ok(line[prev_line_index..line_index].to_vec());
            }

            column_index += 1;
            prev_line_index = line_index + 1;
            if line_index >= line.len() {
                break;
            }
        }

        Err(IoError::Other(format!("Column index {} out of bounds", col_index)))
    }

    /// Evaluate a condition against a line
    fn evaluate_condition(&self, line: &[u8]) -> Result<bool, IoError> {
        match self.condition.as_ref().unwrap() {
            Condition::Compare { column, op, value } => {
                let col_value = self.get_column_value(line, column)?;
                Ok(self.compare(&col_value, *op, value))
            }
            Condition::And(left, right) => {
                // Evaluate left condition
                let mut temp_filter = self.clone_without_reader();
                temp_filter.condition = Some((**left).clone());
                let left_result = temp_filter.evaluate_condition(line)?;

                if !left_result {
                    return Ok(false);
                }

                // Evaluate right condition
                temp_filter.condition = Some((**right).clone());
                let right_result = temp_filter.evaluate_condition(line)?;

                Ok(right_result)
            }
            Condition::Or(left, right) => {
                // Evaluate left condition
                let mut temp_filter = self.clone_without_reader();
                temp_filter.condition = Some((**left).clone());
                let left_result = temp_filter.evaluate_condition(line)?;

                if left_result {
                    return Ok(true);
                }

                // Evaluate right condition
                temp_filter.condition = Some((**right).clone());
                let right_result = temp_filter.evaluate_condition(line)?;

                Ok(right_result)
            }
        }
    }

    /// Helper to clone filter state without the reader
    fn clone_without_reader(&self) -> Self {
        Filter {
            reader: Box::new(DummyReader),
            buffer: self.buffer.clone(),
            header: self.header.clone(),
            condition: self.condition.clone(),
            delimiter: self.delimiter,
        }
    }

    /// Compare two byte slices based on the operator
    fn compare(&self, left: &[u8], op: ComparisonOp, right: &[u8]) -> bool {
        match op {
            ComparisonOp::Eq => left == right,
            ComparisonOp::Ne => left != right,
            ComparisonOp::Lt => self.compare_numeric_or_lexical(left, right).is_lt(),
            ComparisonOp::Le => self.compare_numeric_or_lexical(left, right).is_le(),
            ComparisonOp::Gt => self.compare_numeric_or_lexical(left, right).is_gt(),
            ComparisonOp::Ge => self.compare_numeric_or_lexical(left, right).is_ge(),
        }
    }

    /// Compare values, trying numeric comparison first, then lexical
    fn compare_numeric_or_lexical(&self, left: &[u8], right: &[u8]) -> std::cmp::Ordering {
        // Try to parse as numbers
        let left_str = String::from_utf8_lossy(left);
        let right_str = String::from_utf8_lossy(right);

        if let (Ok(left_num), Ok(right_num)) = (left_str.parse::<f64>(), right_str.parse::<f64>()) {
            left_num.partial_cmp(&right_num).unwrap_or(std::cmp::Ordering::Equal)
        } else {
            left.cmp(right)
        }
    }
}

/// Dummy reader for cloning filter state
struct DummyReader;

impl LineReader for DummyReader {
    fn next_line(&mut self, _buffer: &mut [u8]) -> Result<ops::Range<usize>, IoError> {
        Err(IoError::EndOfFile)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;
    use crate::io::file_reader::FileReader;

    fn setup(content: &[u8], condition: Option<Condition>) -> Filter {
        let mut file = NamedTempFile::new().expect("Failed to create temp file");
        file.write(content).expect("Failed to write to temp file");
        let path = file.path().to_str().unwrap();
        let reader = FileReader::open(path).unwrap();
        let mut filter = Filter::new(Box::new(reader), b',').unwrap();
        filter.read_header().unwrap();
        if let Some(cond) = condition {
            filter.set_condition(cond);
        }
        filter
    }

    #[test]
    fn test_filter_gt() {
        let content = b"a,b\n5,10\n15,8\n3,12\n";
        let mut buffer = vec![0u8; 50];
        let condition = Condition::Compare {
            column: b"a".to_vec(),
            op: ComparisonOp::Gt,
            value: b"4".to_vec(),
        };
        let mut filter = setup(content, Some(condition));

        // Should return rows where a > 4 (5,10 and 15,8)
        let range = filter.next_line(&mut buffer).unwrap();
        assert_eq!(&buffer[range], b"5,10");

        let range = filter.next_line(&mut buffer).unwrap();
        assert_eq!(&buffer[range], b"15,8");

        assert!(matches!(filter.next_line(&mut buffer), Err(IoError::EndOfFile)));
    }

    #[test]
    fn test_filter_eq() {
        let content = b"name,age\nAlice,30\nBob,25\nAlice,35\n";
        let mut buffer = vec![0u8; 50];
        let condition = Condition::Compare {
            column: b"name".to_vec(),
            op: ComparisonOp::Eq,
            value: b"Alice".to_vec(),
        };
        let mut filter = setup(content, Some(condition));

        let range = filter.next_line(&mut buffer).unwrap();
        assert_eq!(&buffer[range], b"Alice,30");

        let range = filter.next_line(&mut buffer).unwrap();
        assert_eq!(&buffer[range], b"Alice,35");

        assert!(matches!(filter.next_line(&mut buffer), Err(IoError::EndOfFile)));
    }

    #[test]
    fn test_no_condition() {
        let content = b"x,y\n1,2\n3,4\n";
        let mut buffer = vec![0u8; 50];
        let mut filter = setup(content, None);

        // Should return all rows
        let range = filter.next_line(&mut buffer).unwrap();
        assert_eq!(&buffer[range], b"1,2");

        let range = filter.next_line(&mut buffer).unwrap();
        assert_eq!(&buffer[range], b"3,4");

        assert!(matches!(filter.next_line(&mut buffer), Err(IoError::EndOfFile)));
    }

    #[test]
    fn test_filter_and() {
        let content = b"a,b\n5,10\n15,8\n3,12\n";
        let mut buffer = vec![0u8; 50];
        let condition = Condition::And(
            Box::new(Condition::Compare {
                column: b"a".to_vec(),
                op: ComparisonOp::Gt,
                value: b"4".to_vec(),
            }),
            Box::new(Condition::Compare {
                column: b"b".to_vec(),
                op: ComparisonOp::Gt,
                value: b"9".to_vec(),
            }),
        );
        let mut filter = setup(content, Some(condition));

        // Should return only rows where a > 4 AND b > 9 (5,10)
        let range = filter.next_line(&mut buffer).unwrap();
        assert_eq!(&buffer[range], b"5,10");

        assert!(matches!(filter.next_line(&mut buffer), Err(IoError::EndOfFile)));
    }
}
