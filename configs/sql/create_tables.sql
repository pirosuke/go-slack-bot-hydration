drop table if exists hydrations;
create table hydrations(
    id serial
    ,username varchar(255) not null
    ,drink varchar(255) not null
    ,amount int not null
    ,modified timestamp not null
    ,primary key (id)
)
;