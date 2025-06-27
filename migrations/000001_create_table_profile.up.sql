create table profile (id uuid primary key, 
    login varchar not null, 
    password varchar not null, 
    balance int not null,
    withdrawn int not null,
    CONSTRAINT login_unique UNIQUE(login));

    