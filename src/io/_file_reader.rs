// use std::cmp::min;
// use std::fs::File;
// use std::os::unix::fs::FileExt;
// use crate::io::error::IoError;
//
// const BUFFER_SIZE: usize = 1024 * 1024; // 1 MB
//
// pub struct TabularFileReader {
//     path: String,
//     file: File,
//     file_offset: u64,
//     buffer: [u8; BUFFER_SIZE],
//     buffer_pointer: usize,
//     eof_pointer: usize,
//     eol_flag: bool,
//     delimiter: u8,
// }
//
// pub struct RowIter<'a> {
//     reader: &'a mut TabularFileReader,
// }
//
// pub struct CellIter<'a> {
//     reader: &'a mut TabularFileReader,
// }
//
// impl<'a> Iterator for CellIter<'a> {
//     type Item = &'a str;
//
//     fn next(&mut self) -> Option<Self::Item> {
//         if self.pos >= self.row.len() {
//             return None;
//         }
//         let bytes = self.row.as_bytes();
//         let start = self.pos;
//         while self.pos < self.row.len() && bytes[self.pos] != self.delimiter {
//             self.pos += 1;
//         }
//         let end = self.pos;
//         self.pos += 1; // skip delimiter
//         Some(&self.row[start..end])
//     }
// }
//
// impl<'a> Iterator for RowIter<'a> {
//     type Item = Result<&'a CellIter<'a>, IoError>;
//
//     fn next(&mut self) -> Option<Self::Item> {
//         if !self.reader.is_eof() {
//             None
//         } else {
//             match self.reader.next_col() {
//                 Ok(row) => Some(Ok(&CellIter {
//                     reader: self.reader
//                 })),
//                 Err(e) => Some(Err(e)),
//             }
//         }
//     }
// }
//
// impl TabularFileReader {
//     fn derive_delimiter(path: &str ) -> u8 {
//         // Simple heuristic based on file extension
//         // TODO: more sophisticated methods may be needed
//         if path.ends_with(".tsv") || path.ends_with(".tab") {
//             b'\t'
//         } else if path.ends_with(".csv") {
//             b','
//         } else if path.ends_with(".psv") {
//             b'|'
//         } else if path.ends_with(".scsv") {
//             b';'
//         } else if path.ends_with(".ssv") {
//             b' '
//         } else {
//             b'\t'
//         }
//     }
//
//     pub fn new(path: String) -> anyhow::Result<Self> {
//         let delimiter = Self::derive_delimiter(path.as_str());
//         let file = File::open(&path)?;
//         Ok(TabularFileReader {
//             path,
//             file,
//             file_offset: 0,
//             buffer: [0; BUFFER_SIZE],
//             buffer_pointer: 0,
//             eof_pointer: BUFFER_SIZE,
//             eol_flag: false,
//             delimiter,
//         })
//     }
//
//     fn load_buffer(&mut self, current_pointer: usize) -> Result<(), IoError> {
//         if self.eof_pointer < BUFFER_SIZE {
//             // ┬──┬ ノ(ò_óノ)
//             return Err(IoError::EndOfFile { path: self.path.as_str() });
//         }
//
//         let available_buffer = &mut self.buffer[current_pointer..];
//
//         if available_buffer.is_empty() {
//             // ༼ つ ◕_◕ ༽つ
//             // Buffer is full, cannot read more data
//             return Err(IoError::LineTooLong { path: self.path.as_str() });
//         }
//
//         let bytes_read = self.file.read_at(available_buffer, self.file_offset)?;
//         self.file_offset += bytes_read as u64;
//         self.eof_pointer = self.buffer_pointer + bytes_read;
//         Ok(())
//     }
//
//     fn move_unread_data_to_front(&mut self, current_pointer: &mut usize) {
//         *current_pointer -= self.buffer_pointer;
//         self.buffer.copy_within(self.buffer_pointer.., 0);
//         self.buffer_pointer = 0;
//     }
//
//     fn next_col(&mut self) -> Result<&[u8], IoError> {
//         let mut found: Option<&[u8]> = None;
//         for mut i in self.buffer_pointer..min(self.buffer.len(), self.eof_pointer) {
//             // TODO: handle values enclosed in quotes
//             if i >= self.buffer.len() - 2 && self.eof_pointer == self.buffer.len() {
//                 self.move_unread_data_to_front(&mut i);
//                 self.load_buffer(i)?;
//             }
//             if self.buffer[i] == self.delimiter || self.buffer[i] == b'\n' || self.buffer[i] == b'\r' {
//                 self.eol_flag = self.buffer[i] == b'\n' || self.buffer[i] == b'\r';
//                 let slice = &self.buffer[self.buffer_pointer..i];
//                 self.buffer_pointer = i + 1;
//                 if self.buffer[i] == b'\r' && i + 1 < self.buffer.len() - 1 && self.buffer[i + 1] == b'\n' {
//                     self.buffer_pointer += 1;
//                 }
//                 return Ok(slice)
//             }
//         }
//         self.eol_flag = true;
//         Ok(&self.buffer[self.buffer_pointer..self.eof_pointer])
//     }
//
//     fn is_eof(&self) -> bool {
//         self.buffer_pointer != self.eof_pointer
//     }
//
//     fn parse_headers(&mut self) -> Result<(), IoError> {
//
//         Ok(())
//     }
//
//     pub fn iter(&mut self) -> RowIter {
//         RowIter { reader: self }
//     }
// }
