create table transactions (
    to_addr binary(20),
    from_addr binary(20),
    hash binary(32),
    block double,
    amount double,
    fee double,
    foreign key (block) references blocks(number) on delete cascade,
    index(to_addr),
    index(from_addr)
);