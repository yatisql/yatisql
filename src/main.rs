mod cli;
mod parser;
mod sql_dialect;
mod query_builder;
mod io;
mod utils;

use cli::build_cli;
use env_logger;
use env_logger::Env;
use clap::ArgMatches;
use crate::io::error::IoError;
use crate::io::traits::BUFFER_SIZE;
use crate::sql_dialect::YatSqlDialect;
use crate::parser::parse_sql;
use crate::query_builder::executor::Builder;

fn run(matches: &ArgMatches) -> anyhow::Result<()> {
    let input = matches.get_one::<String>("input");
    if !input.is_none() {
        log::info!("Input: '{}'", input.unwrap());
    }

    let sql = "\
        SELECT a, b \
        INTO testdata/output.csv \
        FROM testdata/test_table.csv \
        USE INDEX (a,b) \
        WHERE a > b;\
    ";

    let dialect = YatSqlDialect {};
    let _result = parse_sql(&dialect, sql)?;

    for statement in _result {
        let mut executor = statement.build()?;

        loop {
            let mut buffer = vec![0u8; BUFFER_SIZE];
            let line = executor.next_line(&mut buffer);
            match line {
                Ok(range) => {
                    let line_data = &buffer[range];
                    let line_str = String::from_utf8_lossy(line_data);
                    println!("{}", line_str);
                }
                Err(e) => {
                    if let IoError::EndOfFile = e {
                        break;
                    } else {
                        return Err(anyhow::anyhow!("{}", e));
                    }
                }
            }
        }
    }

    Ok(())
}

fn main() -> anyhow::Result<()> {
    // RUST_LOG can override this value
    // TODO: change to YATSQL_LOG
    env_logger::Builder::from_env(Env::default().default_filter_or("debug")).init();

    let matches = build_cli().get_matches();
    let _ = run(&matches)?;
    Ok(())
}
