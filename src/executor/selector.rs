use std::path::Path;
use csv::{Reader, StringRecord, ReaderBuilder};
use crate::executor::row::{Header, Row};

pub trait Selector: Iterator<Item = Row> {
    fn header(&self) -> Header;
}

pub struct FileReaderSelector<'a> {
    header: Header,
    records: Box<dyn Iterator<Item = anyhow::Result<Row>> + 'a>,
}

impl<'a> FileReaderSelector<'a> {
    pub fn new(reader: &'a mut TabularFileReader) -> anyhow::Result<Self> {
        let header = Header::from_string_record(reader.headers());
        let records = Box::new(
            reader.records().map(
                |r| r.map(
                    |record| Row::from_string_record(&record)
                )
            )
        );
        Ok(Self { header, records })
    }
}

impl<'a> Iterator for FileReaderSelector<'a> {
    type Item = Row;

    fn next(&mut self) -> Option<Self::Item> {
        while let Some(result) = self.records.next() {
            match result {
                Ok(row) => return Some(row),
                Err(_) => continue, // skip errors
            }
        }
        None
    }
}

impl<'a> Selector for FileReaderSelector<'a> {
    fn header(&self) -> Header {
        self.header.clone()
    }
}

pub struct TabularFileReader {
    reader: Reader<std::fs::File>,
    headers: StringRecord,
}

impl TabularFileReader {
    pub fn new_csv<P: AsRef<Path>>(file_path: P) -> anyhow::Result<Self> {
        Self::new(file_path, b',')
    }

    pub fn new_tsv<P: AsRef<Path>>(file_path: P) -> anyhow::Result<Self> {
        Self::new(file_path, b'\t')
    }

    pub fn new<P: AsRef<Path>>(file_path: P, delimiter: u8) -> anyhow::Result<Self> {
        let path = file_path.as_ref();
        let mut builder = ReaderBuilder::new();
        builder.delimiter(delimiter);

        let mut reader = builder.from_path(path)?;
        let headers = reader.headers()?.clone();
        Ok(TabularFileReader { reader, headers })
    }

    pub fn headers(&self) -> &StringRecord {
        &self.headers
    }

    pub fn records(&mut self) -> impl Iterator<Item = anyhow::Result<StringRecord>> + '_ {
        self.reader.records().map(|result| result.map_err(anyhow::Error::from))
    }
}

impl Iterator for TabularFileReader {
    type Item = anyhow::Result<StringRecord>;

    fn next(&mut self) -> Option<Self::Item> {
        self.reader
            .records()
            .next()
            .map(|result| result.map_err(anyhow::Error::from))
    }
}
