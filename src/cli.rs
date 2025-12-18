use clap::{Command, Arg};

pub fn build_cli() -> Command {
    Command::new("yatisql")
        .version("0.1.0")
        .about("")
        .arg(Arg::new("input").help("Input sql query or file").required(false))
}
