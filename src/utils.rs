pub fn next_delimiter_pos(buffer: &[u8], start: usize, delimiter: u8) -> usize {
    for i in start..buffer.len() {
        if buffer[i] == delimiter {
            return i;
        }
    }
    buffer.len()
}

pub fn derive_delimiter(path: &str ) -> u8 {
    // Simple heuristic based on file extension
    // TODO: more sophisticated methods may be needed
    if path.ends_with(".tsv") || path.ends_with(".tab") {
        b'\t'
    } else if path.ends_with(".csv") {
        b','
    } else if path.ends_with(".psv") {
        b'|'
    } else if path.ends_with(".scsv") {
        b';'
    } else if path.ends_with(".ssv") {
        b' '
    } else {
        b'\t'
    }
}

pub fn col_name_to_index(column_name: &[u8]) -> Option<usize> {
    let mut index: Option<usize> = None;
    for &ch in column_name.iter() {
        if ch >= b'A' && ch <= b'Z' {
            index = Some(
                index.unwrap_or(0) * (b'Z' - b'A' + 1) as usize
                    + (ch - b'A' + 1) as usize
            );
        } else {
            return None;
        }
    }
    if let Some(i) = index {
        return Some(i - 1)
    }
    index
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test() {
        assert_eq!(col_name_to_index(b"A"), Some(0));
        assert_eq!(col_name_to_index(b"B"), Some(1));
        assert_eq!(col_name_to_index(b"Z"), Some(25));
        assert_eq!(col_name_to_index(b"AA"), Some(26));
        assert_eq!(col_name_to_index(b"AB"), Some(27));
        assert_eq!(col_name_to_index(b"AZ"), Some(51));
        assert_eq!(col_name_to_index(b"BA"), Some(52));
        assert_eq!(col_name_to_index(b"AAA"), Some(702));
        assert_eq!(col_name_to_index(b"AAB"), Some(703));
        assert_eq!(col_name_to_index(b"a"), None);
        assert_eq!(col_name_to_index(b"aA"), None);
        assert_eq!(col_name_to_index(b"A1"), None);
        assert_eq!(col_name_to_index(b""), None);
    }
}
