use sqlparser::ast::{Expr, ObjectName, ObjectNamePart, SelectItem, SetExpr, Statement, TableFactor, TableWithJoins};
use crate::executor::row::Header;
use crate::executor::selector::{FileReaderSelector, Selector, TabularFileReader};

pub fn execute(statement: Statement) -> anyhow::Result<()> {
    match statement {
        Statement::Query(query) => {
            log::debug!("Executing query: {:#?}", query);
            match *query.body {
                SetExpr::Select(select) => {
                    log::debug!("Select statement: {:#?}", select);
                    execute_select(&select)?;
                    // log::info!("Selected projection: {:#?}", select.projection);
                },
                _ => return Err(anyhow::anyhow!("Only SELECT statements are supported.")),
            }
            Ok(())
        },
        _ => Err(anyhow::anyhow!("Only SELECT statements are supported.")),
    }
}

fn execute_select(select: &sqlparser::ast::Select) -> anyhow::Result<()> {
    log::debug!("Executing SELECT: {:#?}", select);
    let mut selected_columns: Vec<String> = Vec::new();
    select.projection.iter().for_each(|item| {
        log::debug!("Projection item: {:#?}", item);
        match item {
            SelectItem::UnnamedExpr(expression) => {
                match expression {
                    Expr::Identifier(identifier) => {
                        selected_columns.push(identifier.value.clone())
                    }
                    Expr::CompoundIdentifier(_) => {}
                    Expr::CompoundFieldAccess { .. } => {}
                    Expr::JsonAccess { .. } => {}
                    Expr::IsFalse(_) => {}
                    Expr::IsNotFalse(_) => {}
                    Expr::IsTrue(_) => {}
                    Expr::IsNotTrue(_) => {}
                    Expr::IsNull(_) => {}
                    Expr::IsNotNull(_) => {}
                    Expr::IsUnknown(_) => {}
                    Expr::IsNotUnknown(_) => {}
                    Expr::IsDistinctFrom(_, _) => {}
                    Expr::IsNotDistinctFrom(_, _) => {}
                    Expr::IsNormalized { .. } => {}
                    Expr::InList { .. } => {}
                    Expr::InSubquery { .. } => {}
                    Expr::InUnnest { .. } => {}
                    Expr::Between { .. } => {}
                    Expr::BinaryOp { .. } => {}
                    Expr::Like { .. } => {}
                    Expr::ILike { .. } => {}
                    Expr::SimilarTo { .. } => {}
                    Expr::RLike { .. } => {}
                    Expr::AnyOp { .. } => {}
                    Expr::AllOp { .. } => {}
                    Expr::UnaryOp { .. } => {}
                    Expr::Convert { .. } => {}
                    Expr::Cast { .. } => {}
                    Expr::AtTimeZone { .. } => {}
                    Expr::Extract { .. } => {}
                    Expr::Ceil { .. } => {}
                    Expr::Floor { .. } => {}
                    Expr::Position { .. } => {}
                    Expr::Substring { .. } => {}
                    Expr::Trim { .. } => {}
                    Expr::Overlay { .. } => {}
                    Expr::Collate { .. } => {}
                    Expr::Nested(_) => {}
                    Expr::Value(_) => {}
                    Expr::Prefixed { .. } => {}
                    Expr::TypedString(_) => {}
                    Expr::Function(_) => {}
                    Expr::Case { .. } => {}
                    Expr::Exists { .. } => {}
                    Expr::Subquery(_) => {}
                    Expr::GroupingSets(_) => {}
                    Expr::Cube(_) => {}
                    Expr::Rollup(_) => {}
                    Expr::Tuple(_) => {}
                    Expr::Struct { .. } => {}
                    Expr::Named { .. } => {}
                    Expr::Dictionary(_) => {}
                    Expr::Map(_) => {}
                    Expr::Array(_) => {}
                    Expr::Interval(_) => {}
                    Expr::MatchAgainst { .. } => {}
                    Expr::Wildcard(_) => {}
                    Expr::QualifiedWildcard(_, _) => {}
                    Expr::OuterJoin(_) => {}
                    Expr::Prior(_) => {}
                    Expr::Lambda(_) => {}
                    Expr::MemberOf(_) => {}
                }
            }
            SelectItem::ExprWithAlias { .. } => {}
            SelectItem::QualifiedWildcard(_, _) => {}
            SelectItem::Wildcard(_) => {}
        }
    });
    log::debug!("Selected columns: {:#?}", selected_columns);
    let mut table_name: Option<String> = None;
    for select_from in &select.from {
        match &select_from.relation {
            TableFactor::Table { name, .. } => {
                // for part in &name.0 {
                //     log::debug!("Table name part: {:#?}", part);
                //
                // }
                let name = &name.0[0];
                match name {
                    ObjectNamePart::Identifier(name) => {
                        table_name = Option::from(name.value.to_string());
                        log::debug!("Table in FROM clause: {:#?}", table_name);
                    }
                    ObjectNamePart::Function(_) => {}
                }
            }
            TableFactor::Derived { .. } => {}
            TableFactor::TableFunction { .. } => {}
            TableFactor::Function { .. } => {}
            TableFactor::UNNEST { .. } => {}
            TableFactor::JsonTable { .. } => {}
            TableFactor::OpenJsonTable { .. } => {}
            TableFactor::NestedJoin { .. } => {}
            TableFactor::Pivot { .. } => {}
            TableFactor::Unpivot { .. } => {}
            TableFactor::MatchRecognize { .. } => {}
            TableFactor::XmlTable { .. } => {}
            TableFactor::SemanticView { .. } => {}
        }
    }
    if table_name.is_none() {
        return Err(anyhow::anyhow!("No table found in FROM clause."));
    }
    let table_name = table_name.unwrap();
    log::info!("Table to query: {}", table_name);

    let reader = TabularFileReader::new_csv(table_name)?;
    let selector = FileReaderSelector::new(
        reader
    )?;

    let old_header = selector.header();
    let new_header = Header::new(selected_columns.clone());
    log::debug!("Old header: {:#?}", old_header);
    log::debug!("New header: {:#?}", new_header);

    for header_item in new_header.iter() {
        print!("{}\t", header_item)
    }
    print!("\n");

    for row in selector {
        let row = row.transform(&old_header, &new_header)?;
        for row_item in row.iter() {
            print!("{}\t", row_item)
        }
        print!("\n");
    }
    Ok(())
}

