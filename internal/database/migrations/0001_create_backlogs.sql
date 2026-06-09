CREATE TABLE backlogs (
    id          VARCHAR2(36)   NOT NULL,
    title       VARCHAR2(200)  NOT NULL,
    description VARCHAR2(1000) NOT NULL,
    created_at  TIMESTAMP      DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT pk_backlogs PRIMARY KEY (id)
)
