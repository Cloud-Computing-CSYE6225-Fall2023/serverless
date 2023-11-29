import os
import uuid
import zipfile
import shutil
from google.cloud import storage
from datetime import datetime, timedelta


def extract_zip(zip_file_path, extract_folder):
    with zipfile.ZipFile(zip_file_path, 'r') as zip_ref:
        zip_ref.extractall(extract_folder)


def get_extracted_size(extract_folder):
    total_size = 0
    for foldername, subfolders, filenames in os.walk(extract_folder):
        for filename in filenames:
            file_path = os.path.join(foldername, filename)
            total_size += os.path.getsize(file_path)
    return total_size


def write_to_blob(file_resp, file_path):
    try:
        bucket_name = os.environ["bucket_name"]
        gcp_project_id = os.environ["project_id"]

        credentials = {
            "type": os.environ["type"],
            "project_id": os.environ["project_id"],
            "private_key_id": os.environ["private_key_id"],
            "private_key": os.environ["private_key"],
            "client_email": os.environ["client_email"],
            "client_id": os.environ["client_id"],
            "auth_uri": os.environ["auth_uri"],
            "token_uri": os.environ["token_uri"],
            "auth_provider_x509_cert_url": os.environ["auth_provider_x509_cert_url"],
            "client_x509_cert_url": os.environ["client_x509_cert_url"],
            "universe_domain": os.environ["universe_domain"]
        }

        storage_client = storage.Client.from_service_account_info(credentials)
        bucket = storage_client.bucket(bucket_name)
        blob = bucket.blob(file_path)

        randomId = str(uuid.uuid4())
        destinationPath = "./" + randomId + ".zip"
        with open(destinationPath, 'wb') as file:
            for chunk in file_resp.iter_content(chunk_size=8192):
                if chunk:
                    file.write(chunk)
            file.seek(0)

            extract_to = "./" + randomId
            extract_zip(destinationPath, extract_to)
            if get_extracted_size(extract_to) == 0:
                return {
                    "statusCode": 500,
                    "msg": "Empty Zip file"
                }

            blob.upload_from_filename(destinationPath, content_type='application/zip')
            expiration_time = datetime.utcnow() + timedelta(minutes=10)

            file.close()

        os.remove(destinationPath)
        shutil.rmtree(extract_to)

        return {
            "statusCode": 200,
            "path": blob.generate_signed_url(expiration_time),
            "msg": "Successfully uploaded file to GCS"
        }

    except Exception as err:
        return {
            "statusCode": 500,
            "msg": "Error Uploading file to GCS: " + str(err)
        }
