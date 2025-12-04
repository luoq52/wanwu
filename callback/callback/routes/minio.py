import logging
import posixpath
from urllib.parse import urljoin

from flask import jsonify, request

from callback.services import minio as minio_service
from configs.config import config

from . import callback_bp


@callback_bp.route("/upload", methods=["POST"])
def upload_file():
    """
    上传文件到 MinIO 存储
    ---
    summary: 上传文件到 MinIO 存储
    tags:
      - Minio
    requestBody:
      required: true
      content:
        multipart/form-data:
          schema:
            type: object
            required:
              - file
            properties:
              file:
                type: string
                format: binary
                description: 需要上传的文件
              bucket_name:
                type: string
                description: 目标存储桶名称 (默认为 callback)
                default: callback
              file_name:
                type: string
                description: (可选) 重写保存的文件名
    responses:
      '200':
        description: 上传成功
        content:
          application/json:
            schema:
              type: object
              properties:
                download_link:
                  type: string
                  description: 文件的下载链接
                  example: "http://base-url/callback/filename.jpg"
      '500':
        description: 服务器内部错误
        content:
          application/json:
            schema:
              type: object
              properties:
                error:
                  type: string
                  example: "Failed to upload file to Minio."
    """
    try:
        default_bucket_name = config.callback_cfg["MINIO"]["BUCKET_NAME"]
        base_url = config.callback_cfg["URL"]["MINIO_DOWNLOAD"]
        # port = response_data['data']['massAccessPort']
        public_minio_download_url = base_url
        # 从请求中获取文件
        uploaded_file = request.files["file"]
        original_filename = uploaded_file.filename

        # ------ Get bucket name and file name from request data ------
        bucket_name = request.form.get("bucket_name", default_bucket_name)
        overwrite_file_name = request.form.get("file_name", None)
        # ------------------------------------------------------------

        uploaded_file_name, timestamp = minio_service.upload_file_to_minio(
            uploaded_file, original_filename, bucket_name, overwrite_file_name
        )
        if uploaded_file_name:
            download_link = posixpath.join(
                public_minio_download_url, bucket_name, timestamp, uploaded_file_name
            )
            logging.info(f"File uploaded successfully, download link: {download_link}")

            return jsonify({"download_link": download_link})
        else:
            return jsonify({"error": "Failed to upload file to Minio."}), 500
    except Exception as e:
        logging.error(f"Error in upload_file endpoint: {e}")
        return jsonify({"error": str(e)}), 500
