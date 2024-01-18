CREATE TABLE `token_balances` (
                                  `block_number` bigint(20) unsigned DEFAULT NULL COMMENT 'Block number containing the transaction',
                                  `block_timestamp` datetime(3) DEFAULT NULL COMMENT 'Block timestamp containing the transaction',
                                  `tick` varchar(255) NOT NULL COMMENT 'Token tick',
                                  `wallet_address` varchar(42) NOT NULL COMMENT 'Address of owner',
                                  `total_supply` decimal(38, 0) DEFAULT NULL COMMENT 'Max supply',
                                  `amount` decimal(38, 0) DEFAULT NULL COMMENT 'The balance of wallet balance at the corresponding block height',
                                  PRIMARY KEY (`tick`, `wallet_address`)
);
