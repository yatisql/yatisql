use std::{fmt, io};

#[derive(Debug)]
pub enum IoError {
    Io(io::Error),
    EndOfFile,
    LineTooLong,
    ColumnNotFound(String),
    Other(String),
}

impl fmt::Display for IoError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            IoError::Io(e) => write!(f, "I/O error: {}", e),
            IoError::EndOfFile {  } => write!(f, "End of file reached",),
            IoError::LineTooLong {  } => write!(f, "Row is too long"),
            IoError::ColumnNotFound(msg) => write!(f, "Column not found: {}", msg),
            IoError::Other(msg) => write!(f, "{}", msg),
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
