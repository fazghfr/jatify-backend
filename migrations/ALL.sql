-- Created by Redgate Data Modeler (https://datamodeler.redgate-platform.com)
-- Last modification date: 2026-02-21 06:31:05.303

-- tables
-- Table: applications
CREATE TABLE applications (
    id int  NOT NULL,
    resume_id int  NOT NULL,
    job_id int  NOT NULL,
    text text  NOT NULL,
    created_at timestamp  NOT NULL,
    updated_at timestamp  NOT NULL,
    deleted_at timestamp  NOT NULL,
    status_id int  NOT NULL,
    user_id int  NOT NULL,
    uuid Varchar(36)  NOT NULL,
    CONSTRAINT applications_pk PRIMARY KEY (id)
);

-- Table: jobs
CREATE TABLE jobs (
    id int  NOT NULL,
    company varchar(256)  NOT NULL,
    position varchar(256)  NOT NULL,
    description text  NOT NULL,
    created_at timestamp  NOT NULL,
    updated_at timestamp  NOT NULL,
    user_id int  NOT NULL,
    uuid Varchar(255)  NOT NULL,
    CONSTRAINT jobs_pk PRIMARY KEY (id)
);

-- Table: resumes
CREATE TABLE resumes (
    id int  NOT NULL,
    filepath text  NOT NULL,
    user_id int  NOT NULL,
    uuid Varchar(36)  NOT NULL,
    CONSTRAINT resumes_pk PRIMARY KEY (id)
);

-- Table: status_history
CREATE TABLE status_history (
    id int  NOT NULL,
    application_id int  NOT NULL,
    status_id int  NOT NULL,
    created_at timestamp  NOT NULL,
    CONSTRAINT id PRIMARY KEY (id)
);

-- Table: statuses
CREATE TABLE statuses (
    id int  NOT NULL,
    text varchar(50)  NOT NULL,
    CONSTRAINT statuses_pk PRIMARY KEY (id)
);

-- Table: users
CREATE TABLE users (
    id int  NOT NULL,
    username varchar(64)  NOT NULL,
    name varchar(64)  NOT NULL,
    phone varchar(64)  NOT NULL,
    email varchar(128)  NOT NULL,
    password varchar(64)  NOT NULL,
    uuid varchar(36)  NOT NULL,
    CONSTRAINT users_pk PRIMARY KEY (id)
);

-- foreign keys
-- Reference: applications_jobs (table: applications)
ALTER TABLE applications ADD CONSTRAINT applications_jobs
    FOREIGN KEY (job_id)
    REFERENCES jobs (id)  
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

-- Reference: applications_resumes (table: applications)
ALTER TABLE applications ADD CONSTRAINT applications_resumes
    FOREIGN KEY (resume_id)
    REFERENCES resumes (id)  
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

-- Reference: applications_statuses (table: applications)
ALTER TABLE applications ADD CONSTRAINT applications_statuses
    FOREIGN KEY (status_id)
    REFERENCES statuses (id)  
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

-- Reference: applications_users (table: applications)
ALTER TABLE applications ADD CONSTRAINT applications_users
    FOREIGN KEY (user_id)
    REFERENCES users (id)  
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

-- Reference: jobs_users (table: jobs)
ALTER TABLE jobs ADD CONSTRAINT jobs_users
    FOREIGN KEY (user_id)
    REFERENCES users (id)  
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

-- Reference: resumes_users (table: resumes)
ALTER TABLE resumes ADD CONSTRAINT resumes_users
    FOREIGN KEY (user_id)
    REFERENCES users (id)  
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

-- Reference: status_history_applications (table: status_history)
ALTER TABLE status_history ADD CONSTRAINT status_history_applications
    FOREIGN KEY (application_id)
    REFERENCES applications (id)  
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

-- Reference: status_history_statuses (table: status_history)
ALTER TABLE status_history ADD CONSTRAINT status_history_statuses
    FOREIGN KEY (status_id)
    REFERENCES statuses (id)  
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

-- sequences
-- Sequence: applications_seq
CREATE SEQUENCE applications_seq
      INCREMENT BY 1
      NO MINVALUE
      NO MAXVALUE
      START WITH 1
      NO CYCLE
;

-- Sequence: jobs_seq
CREATE SEQUENCE jobs_seq
      INCREMENT BY 1
      NO MINVALUE
      NO MAXVALUE
      START WITH 1
      NO CYCLE
;

-- Sequence: resumes_seq
CREATE SEQUENCE resumes_seq
      INCREMENT BY 1
      NO MINVALUE
      NO MAXVALUE
      START WITH 1
      NO CYCLE
;

-- Sequence: statuses_seq
CREATE SEQUENCE statuses_seq
      INCREMENT BY 1
      NO MINVALUE
      NO MAXVALUE
      START WITH 1
      NO CYCLE
;

-- End of file.

