use sqlparser::ast::{Expr, ObjectNamePart, Select, SelectItem, SetExpr, Statement, TableFactor};
use crate::io::file_reader::FileReader;
use crate::io::file_writer::FileWriter;
use crate::io::selector::Selector;
use crate::io::traits::LineReader;
use crate::utils::derive_delimiter;

pub trait Builder {
    type Output;
    fn build(self) -> anyhow::Result<Self::Output>;
}

impl Builder for Statement {
    type Output = Box<dyn LineReader>;
    fn build(self) -> anyhow::Result<Self::Output> {
        match self {
            Statement::Query(query) => {
                log::debug!("query: {:#?}", query);
                match *query.body {
                    SetExpr::Select(select) => {
                        log::debug!("Select statement: {:#?}", select);
                        select.build()
                    },
                    _ => Err(anyhow::anyhow!("Only SELECT statements are supported.")),
                }
            },
            _ => Err(anyhow::anyhow!("Only SELECT statements are supported.")),
        }
    }
}

impl Builder for Select {
    type Output = Box<dyn LineReader>;
    fn build(self) -> anyhow::Result<Self::Output> {
        log::debug!("SELECT: {:#?}", self);
        let mut selected_columns: Vec<&[u8]> = Vec::new();
        self.projection.iter().for_each(|item| {
            log::debug!("Projection item: {:#?}", item);
            match item {
                SelectItem::UnnamedExpr(expression) => {
                    match expression {
                        Expr::Identifier(identifier) => {
                            selected_columns.push(identifier.value.as_bytes())
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
        for select_from in self.from {
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
        let mut into_tablenames: Vec<String> = Vec::new();
        if let Some(select_into) = self.into {
            for part in select_into.name.0 {
                log::debug!("INTO clause table name part: {:#?}", part);
                match part {
                    ObjectNamePart::Identifier(name) => {
                        into_tablenames.push(name.value);
                    }
                    ObjectNamePart::Function(_) => {}
                }
            }
        }

        if table_name.is_none() {
            return Err(anyhow::anyhow!("No table found in FROM clause."));
        }
        let table_name = table_name.unwrap();
        log::info!("Table to query: {}", table_name);

        let reader = FileReader::open(&table_name)?;
        let mut selector = Selector::new(
            Box::new(reader),
            derive_delimiter(&table_name)
        )?;
        selector.read_header()?;
        selector.set_rules(&selected_columns)?;

        let mut result: Box<dyn LineReader> = Box::new(selector);

        for into_tablename in into_tablenames {
            result = Box::new(
                FileWriter::new(
                    result,
                    &into_tablename
                )?
            );
        }

        Ok(result)
    }
}

