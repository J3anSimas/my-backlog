CREATE TABLE mbl_users (
    id            VARCHAR2(36)  NOT NULL,
    name          VARCHAR2(100) NOT NULL,  -- mirrors user.MaxNameLength
    email         VARCHAR2(254) NOT NULL,  -- mirrors user.MaxEmailLength; RFC 5321
    password_hash VARCHAR2(60)  NOT NULL,  -- bcrypt output fixo em 60 chars
    created_at    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT pk_mbl_users       PRIMARY KEY (id),
    CONSTRAINT uq_mbl_users_email UNIQUE (email)
)
