select
    hash,
    nonce,
    transaction_index,
    from_address,
    to_address,
    value,
    gas,
    gas_price,
    concat('0x',lower(to_hex(input))),
    receipt_cumulative_gas_used,
    receipt_gas_used,
    receipt_contract_address,
    receipt_root,
    receipt_status,
    block_timestamp,
    block_number,
    block_hash,
    max_fee_per_gas,
    max_priority_fee_per_gas,
    transaction_type,
    receipt_effective_gas_price
from
    (
        (
            select
                hash,
                nonce,
                transaction_index,
                from_address,
                to_address,
                value,
                gas,
                gas_price,
                input,
                receipt_cumulative_gas_used,
                receipt_gas_used,
                receipt_contract_address,
                '' AS receipt_root,
                receipt_status,
                cast(
                to_unixtime(block_timestamp) as bigint
                ) as block_timestamp,
                block_number,
                block_hash,
                max_fee_per_gas,
                max_priority_fee_per_gas,
                transaction_type,
                receipt_effective_gas_price
            from
                avalanche.transactions
            where
                block_number >= 38780167
              and block_number <= 39206439
        ) a
            join (
            select
                tx_hash
            from
                avas.raw
            where
                block_number >= 38780167
              and block_number <= 39206439
              and tick = 'dino'
        ) b on a.hash = b.tx_hash
        );