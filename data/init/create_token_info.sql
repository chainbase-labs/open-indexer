CREATE TABLE IF NOT EXISTS `token_info` (
    `block_timestamp` datetime(3) NOT NULL COMMENT 'Timestamp of the block containing the inscription (matches block_timestamp in transactions table)',
    `block_number` bigint(20) NOT NULL COMMENT 'Block number containing the inscription (matches block_number in transactions table)',
    `tx_index` int(11) NOT NULL COMMENT 'Index of the transaction containing the inscription (matches transaction_index in transactions table)',
    `tx_hash` varchar(66) NOT NULL COMMENT 'Unique identifier of the transaction containing the inscription (matches hash in transactions table)',
    `tick` varchar(255) NOT NULL COMMENT 'Token tick',
    `max_supply` decimal(38, 0) DEFAULT NULL COMMENT 'Max supply',
    `lim` decimal(38, 0) DEFAULT NULL COMMENT 'Limit of each mint',
    `wlim` decimal(38, 0) DEFAULT NULL COMMENT 'Limit of each address can maximum mint',
    `dec` int(11) DEFAULT NULL COMMENT 'Decimal for minimum divie',
    `creator` varchar(42) DEFAULT NULL COMMENT 'Address originating the inscription (matches from_address in transactions table)',
    `minted` decimal(38, 0) DEFAULT '0',
    `holders` decimal(38, 0) DEFAULT '0',
    `txs` decimal(38, 0) DEFAULT '0',
    `updated_timestamp` timestamp(3) NULL DEFAULT NULL,
    `completed_timestamp` timestamp(3) NULL DEFAULT NULL,
    `id` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`tick`)
    );
