#!/bin/bash

# 设置 S3 桶路径
S3_BUCKET_PATH=$1

# 设置本地数据目录
LOCAL_DATA_DIR="./data"

# 确保本地数据目录存在
mkdir -p $LOCAL_DATA_DIR

# 获取 S3 桶中的文件列表并按数字排序
FILES=$(aws s3 ls "$S3_BUCKET_PATH/transactions/" | awk '{print $4}' | sort -n)
echo "$FILES"

# 循环遍历每个文件
for FILE in $FILES; do
    echo "Processing $FILE..."

    # 构建 S3 文件路径和本地文件路径
    S3_TRANSACTION_FILE_PATH="$S3_BUCKET_PATH/transactions/$FILE"
    S3_LOGS_FILE_PATH="$S3_BUCKET_PATH/logs/$FILE"

    LOCAL_TRANSACTION_FILE_PATH="$LOCAL_DATA_DIR/transactions.txt"
    LOCAL_LOG_FILE_PATH="$LOCAL_DATA_DIR/logs.txt"
    if [ -f "$LOCAL_TRANSACTION_FILE_PATH" ]; then
      rm "$LOCAL_TRANSACTION_FILE_PATH"
    if [ -f "$LOCAL_LOG_FILE_PATH" ]; then
      rm "$LOCAL_LOG_FILE_PATH"
    # 拉取数据文件
    aws s3 cp $S3_TRANSACTION_FILE_PATH $LOCAL_TRANSACTION_FILE_PATH
    if aws s3 ls $S3_LOGS_FILE_PATH; then
          aws s3 cp $S3_LOGS_FILE_PATH $LOCAL_LOG_FILE_PATH
    else
          echo "Log file does not exist, creating an empty file."
          touch $LOCAL_LOG_FILE_PATH
    fi

    if [ -f "$LOCAL_TRANSACTION_FILE_PATH" ] && [ -f "$LOCAL_LOG_FILE_PATH" ]; then
        sed -i 's/\"//g' $LOCAL_TRANSACTION_FILE_PATH
        sed -i 's/\"//g' $LOCAL_LOG_FILE_PATH

        echo "index avas ...."
        tidb_db_name=$2 tidb_host=$3 tidb_password=$4 tidb_port=$5 tidb_user=$6 ./indexer --transactions $LOCAL_TRANSACTION_FILE_PATH --logs $LOCAL_LOG_FILE_PATH
    else
        echo "Failed to download $FILE"
    fi
done

echo "All files processed."
