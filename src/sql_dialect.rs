use sqlparser::dialect::{Dialect};
use sqlparser::keywords;
use sqlparser::keywords::Keyword;
use sqlparser::parser::Parser;

#[derive(Debug)]
pub struct YatSqlDialect;

// impl Dialect for YatSqlDialect {
//     fn is_identifier_start(&self, ch: char) -> bool {
//         ch.is_alphabetic() || ch == '_' || ch == '/'
//     }
//
//     fn is_identifier_part(&self, ch: char) -> bool {
//         self.is_identifier_start(ch) || ch.is_digit(10)
//             || ch == '.' || ch == '-' || ch == '\\'
//     }
// }

const RESERVED_FOR_TABLE_ALIAS_YATISQL: &[Keyword] = &[
    Keyword::USE,
    // Keyword::IGNORE,
    // Keyword::FORCE,
    // Keyword::STRAIGHT_JOIN,
];

impl Dialect for YatSqlDialect {
    fn is_delimited_identifier_start(&self, ch: char) -> bool {
        ch == '"' || ch == '`'
    }

    fn is_identifier_start(&self, ch: char) -> bool {
        ch.is_alphabetic() || ch == '_' || ch == '/' || ch == '#' || ch == '@'
    }

    fn is_identifier_part(&self, ch: char) -> bool {
        ch.is_alphabetic()
            || ch.is_ascii_digit()
            || ch == '@'
            || ch == '$'
            || ch == '#'
            || ch == '_'
            || ch == '.'
            || ch == '-'
            || ch == '/'
            || ch == '\\'
    }

    fn is_table_factor_alias(&self, explicit: bool, kw: &Keyword, _parser: &mut Parser) -> bool {
        explicit
            || (!keywords::RESERVED_FOR_TABLE_ALIAS.contains(kw)
            && !RESERVED_FOR_TABLE_ALIAS_YATISQL.contains(kw))
    }

    fn supports_table_hints(&self) -> bool {
        true
    }

    fn supports_unicode_string_literal(&self) -> bool {
        true
    }

    fn supports_group_by_expr(&self) -> bool {
        true
    }

    fn supports_group_by_with_modifier(&self) -> bool {
        true
    }

    fn supports_left_associative_joins_without_parens(&self) -> bool {
        true
    }

    fn supports_connect_by(&self) -> bool {
        true
    }

    fn supports_match_recognize(&self) -> bool {
        true
    }

    fn supports_pipe_operator(&self) -> bool {
        true
    }

    fn supports_start_transaction_modifier(&self) -> bool {
        true
    }

    fn supports_window_function_null_treatment_arg(&self) -> bool {
        true
    }

    fn supports_dictionary_syntax(&self) -> bool {
        true
    }

    fn supports_window_clause_named_window_reference(&self) -> bool {
        true
    }

    fn supports_parenthesized_set_variables(&self) -> bool {
        true
    }

    fn supports_select_wildcard_except(&self) -> bool {
        true
    }

    fn support_map_literal_syntax(&self) -> bool {
        true
    }

    fn allow_extract_custom(&self) -> bool {
        true
    }

    fn allow_extract_single_quotes(&self) -> bool {
        true
    }

    fn supports_create_index_with_clause(&self) -> bool {
        true
    }

    fn supports_explain_with_utility_options(&self) -> bool {
        true
    }

    fn supports_limit_comma(&self) -> bool {
        true
    }

    fn supports_from_first_select(&self) -> bool {
        true
    }

    fn supports_projection_trailing_commas(&self) -> bool {
        true
    }

    fn supports_asc_desc_in_column_definition(&self) -> bool {
        true
    }

    fn supports_try_convert(&self) -> bool {
        true
    }

    fn supports_comment_on(&self) -> bool {
        true
    }

    fn supports_load_extension(&self) -> bool {
        true
    }

    fn supports_named_fn_args_with_assignment_operator(&self) -> bool {
        true
    }

    fn supports_struct_literal(&self) -> bool {
        true
    }

    fn supports_empty_projections(&self) -> bool {
        true
    }

    fn supports_nested_comments(&self) -> bool {
        true
    }

    fn supports_user_host_grantee(&self) -> bool {
        true
    }

    fn supports_string_escape_constant(&self) -> bool {
        true
    }

    fn supports_array_typedef_with_brackets(&self) -> bool {
        true
    }

    fn supports_match_against(&self) -> bool {
        true
    }

    fn supports_set_names(&self) -> bool {
        true
    }

    fn supports_comma_separated_set_assignments(&self) -> bool {
        true
    }

    fn supports_filter_during_aggregation(&self) -> bool {
        true
    }

    fn supports_select_wildcard_exclude(&self) -> bool {
        true
    }

    fn supports_data_type_signed_suffix(&self) -> bool {
        true
    }

    fn supports_interval_options(&self) -> bool {
        true
    }
}
