/**
 * Both functions require access to tables the schema 'shared' which
 * is not accessible for the manager user.
 */
ALTER FUNCTION hook_prohibit_delete_in_use() SECURITY DEFINER;
ALTER FUNCTION hook_check_foreign_key_soft_delete() SECURITY DEFINER;
