import os
import posixpath
from datetime import datetime, timedelta, timezone

from extensions.minio import minio_client
from utils.log import logger


def upload_file_to_minio(
    file_stream, original_filename, bucket_name, overwrite_file_name=None
):
    cst_tz = timezone(timedelta(hours=8))
    timestamp = datetime.now(cst_tz).strftime("%Y%m%d%H%M%S")
    _, file_extension = os.path.splitext(original_filename)
    file_name = original_filename
    if overwrite_file_name:
        file_name = overwrite_file_name + file_extension
    object_name = posixpath.join(timestamp, file_name)
    minio_client.create_public_bucket_if_not_exist(bucket_name)
    minio_client.put_object_from_stream(bucket_name, object_name, file_stream)
    logger.info(f"File '{file_name}' uploaded to Minio bucket '{bucket_name}'")

    return file_name, timestamp
