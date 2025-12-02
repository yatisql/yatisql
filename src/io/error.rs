use std::{fmt, io};
use std::str::Utf8Error;

#[derive(Debug)]
pub enum IoError {
    Io(io::Error),
    EndOfFile { path: String },
    LineTooLong { path: String },
    Utf8Error { path: String, source: Utf8Error },
}

impl fmt::Display for IoError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            IoError::Io(e) => write!(f, "I/O error: {}", e),
            IoError::EndOfFile { path } => write!(f, "End of file reached: {}", path),
            IoError::LineTooLong { path } => write!(f, "Row is too long: {}", path),
            IoError::Utf8Error { path, source } =>
                write!(f, "Utf8 read error: {}, reason: {}", path, source),
        }
    }
}

impl std::error::Error for IoError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            IoError::Io(e) => Some(e),
            _ => None,
        }
    }
}

impl From<io::Error> for IoError {
    fn from(e: io::Error) -> Self {
        IoError::Io(e)
    }
}
