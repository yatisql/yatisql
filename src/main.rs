mod cli;
mod parser;
mod sql_dialect;
mod executor;
mod io;

use cli::build_cli;
use env_logger;
use env_logger::Env;
use clap::ArgMatches;

use crate::sql_dialect::YatSqlDialect;
use crate::parser::parse_sql;
use crate::executor::executor::execute;


fn run(matches: &ArgMatches) -> anyhow::Result<()> {
    let input = matches.get_one::<String>("input");
    if !input.is_none() {
        log::info!("Input: '{}'", input.unwrap());
    }

    let sql = "\
        SELECT a, b \
        FROM testdata/test_table.csv \
        USE INDEX (a,b) \
        WHERE a > b;\
    ";

    let dialect = YatSqlDialect {};
    let _result = parse_sql(&dialect, sql)?;

    for statement in _result {
        execute(statement)?;
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
