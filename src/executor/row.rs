use csv::StringRecord;

#[derive(Debug, Clone)]
pub struct Header(pub Vec<String>);

impl Header {
    pub fn new(data: Vec<String>) -> Self {
        Header(data)
    }

    pub fn iter(&self) -> std::slice::Iter<String> {
        self.0.iter()
    }

    pub fn from_string_record(data: &StringRecord) -> Self {
        Header(data.iter().map(|item| item.to_string()).collect())
    }
}

#[derive(Debug, Clone)]
pub struct Row(pub Vec<String>);

impl Row {
    pub fn new() -> Self {
        Row(Vec::new())
    }

    pub fn from_vec(data: Vec<String>) -> Self {
        Row(data)
    }

    pub fn from_string_record(data: &StringRecord) -> Self {
        Row(data.iter().map(|item| item.to_string()).collect())
    }

    pub fn iter(&self) -> std::slice::Iter<String> {
        self.0.iter()
    }

    pub fn transform(&self, from: &Header, to: &Header) -> anyhow::Result<Row> {
        let mut result: Row = Row::new();
        for to_header in to.iter() {
            let mut found = false;
            for (from_header, value) in from.iter().zip(self.iter()) {
                if from_header == to_header {
                    result.0.push(value.clone());
                    found = true;
                    break;
                }
            }
            if !found {
                return Err(anyhow::anyhow!("Column '{}' not found.", to_header));
            }
        }
        Ok(result)
    }
}


