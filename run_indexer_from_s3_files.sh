#!/bin/bash

# 设置 S3 桶路径
S3_BUCKET_PATH=$1

# 设置本地数据目录
LOCAL_DATA_DIR="./data"

# 确保本地数据目录存在
mkdir -p $LOCAL_DATA_DIR

# 获取 S3 桶中的文件列表并按数字排序
FILES=$(aws s3 ls $S3_BUCKET_PATH | awk '{print $4}' | sort -n)

# 循环遍历每个文件
for FILE in $FILES; do
    echo "Processing $FILE..."

    # 构建 S3 文件路径和本地文件路径
    S3_FILE_PATH="$S3_BUCKET_PATH/$FILE"
    LOCAL_FILE_PATH="$LOCAL_DATA_DIR/transactions.txt"

    # 拉取数据文件
    aws s3 cp $S3_FILE_PATH $LOCAL_FILE_PATH

    echo "download $File ...."
    if [ -f "$LOCAL_FILE_PATH" ]; then
        sed -i 's/\"//g' $LOCAL_FILE_PATH

        echo "index avas ...."
        tidb_db_name=$2 tidb_host=$3 tidb_password=$4 tidb_port=$5 tidb_user=$6 ./indexer --transactions $LOCAL_FILE_PATH --logs ./data/logs.txt
    else
        echo "Failed to download $FILE"
    fi

    rm $LOCAL_FILE_PATH
done

echo "All files processed."
