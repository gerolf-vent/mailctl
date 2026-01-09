/***************************************************************
 * Table for audit logs
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE audit.log (
    ID BIGSERIAL PRIMARY KEY,
    table_name VARCHAR(64) NOT NULL,
    record_id INT NOT NULL,
    operation VARCHAR(10) NOT NULL,  -- INSERT, UPDATE, DELETE
    data_old JSONB,
    data_new JSONB,
    changed_by VARCHAR(256),  -- Application user/session
    changed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_log_table_record ON audit.log(table_name, record_id);
CREATE INDEX idx_audit_log_changed_at ON audit.log(changed_at);
CREATE INDEX idx_audit_log_table_name ON audit.log(table_name);
