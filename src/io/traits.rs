use std::ops;
use crate::io::error::IoError;

pub const BUFFER_SIZE: usize = 1024 * 1024 * 1; // 1 MB buffer

pub trait LineReader {
    fn next_line(&mut self, buffer: &mut [u8]) -> Result<ops::Range<usize>, IoError>;
}

impl<T: LineReader> LineReader for &mut T {
    fn next_line(&mut self, buffer: &mut [u8]) -> Result<ops::Range<usize>, IoError> {
        (**self).next_line(buffer)
    }
}
