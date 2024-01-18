CREATE TABLE IF NOT EXISTS `token_activities` (
                                    `block_timestamp` datetime(3) NOT NULL COMMENT 'Timestamp of the block containing the inscription (matches block_timestamp in transactions table)',
                                    `block_number` bigint(20) NOT NULL COMMENT 'Block number containing the inscription (matches block_number in transactions table)',
                                    `tx_index` int(11) NOT NULL COMMENT 'Index of the transaction containing the inscription (matches transaction_index in transactions table)',
                                    `tx_hash` varchar(66) NOT NULL COMMENT 'Unique identifier of the transaction containing the inscription (matches hash in transactions table)',
                                    `log_index` int(11) NOT NULL COMMENT 'Index of the log within the transaction',
                                    `type` varchar(255) NOT NULL COMMENT 'mint  transfer  burn',
                                    `tick` varchar(255) NOT NULL COMMENT 'Token tick',
                                    `id` varchar(255) NOT NULL COMMENT 'Unique identifier of the inscription',
                                    `amt` decimal(38, 0) DEFAULT NULL COMMENT 'Mint amount',
                                    `from_address` varchar(42) DEFAULT NULL COMMENT 'Address sending the inscription (matches from_address in transactions table)',
                                    `to_address` varchar(42) DEFAULT NULL COMMENT 'Address receiving the inscription (match to_address in transactions table)',
                                    PRIMARY KEY (
                                                 `id`, `log_index`, `tx_index`, `tick`,
                                                 `type`
                                        )
);
