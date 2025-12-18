use sqlparser::dialect::{Dialect};
use sqlparser::parser::{Parser, ParserError};
use log;
use sqlparser::ast::Statement;


pub fn parse_sql(dialect: &dyn Dialect, sql: &str) -> Result<Vec<Statement>, ParserError> {
    let parse_sql_result = Parser::parse_sql(dialect, sql);

    let ast = match parse_sql_result {
        Ok(ast) => ast,
        Err(error) => return Err(error),
    };

    log::debug!("SQL THREE: {:#?}", ast);

    if ast.is_empty() {
        return Err(ParserError::ParserError("Empty SQL statements are not supported.".to_string()))
    }

    if ast.len() != 1 {
        return Err(ParserError::ParserError("Multiple SQL statements are not supported.".to_string()))
    }

    if let Statement::Query(_) = &ast[0] {
    } else {
        return Err(ParserError::ParserError("Only SELECT statements are supported.".to_string()));
    }

    Ok(ast)
}
