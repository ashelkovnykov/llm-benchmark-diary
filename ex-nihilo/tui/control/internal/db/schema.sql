create table records (
    rowid integer primary key,
    count integer not null default 1
) strict;
insert into records (rowid, count) values
    (1,  101),
    (2,  102),
    (3,  103),
    (4,  104),
    (5,  105),
    (6,  106),
    (7,  107),
    (8,  108),
    (9,  109),
    (10, 110),
    (11, 111),
    (12, 112),
    (13, 113),
    (14, 114),
    (15, 115);
