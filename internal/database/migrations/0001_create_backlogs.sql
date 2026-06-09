CREATE TABLE mbl_backlogs (
    id          VARCHAR2(36)   NOT NULL,
    title       VARCHAR2(200)  NOT NULL,  -- mirrors backlog.MaxTitleLength
    description VARCHAR2(1000) NOT NULL, -- mirrors backlog.MaxDescriptionLength
    created_at  TIMESTAMP      DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT pk_mbl_backlogs PRIMARY KEY (id)
)
