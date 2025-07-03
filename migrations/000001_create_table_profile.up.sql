create table profile (id uuid primary key, 
    login varchar not null, 
    password varchar not null, 
    balance numeric(10, 2) not null,
    withdrawn numeric(10, 2) not null,
    CONSTRAINT login_unique UNIQUE(login));

