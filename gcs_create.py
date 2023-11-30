import os
import tempfile
from google.cloud import storage
from datetime import datetime, timedelta


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

        # Create a temporary file
        with tempfile.NamedTemporaryFile(delete=False) as temp_file:
            temp_file_path = temp_file.name

            # Write the content to the temporary file
            for chunk in file_resp.iter_content(chunk_size=8192):
                if chunk:
                    temp_file.write(chunk)

            temp_file.seek(0)
            blob.upload_from_file(temp_file, content_type='application/zip')
            expiration_time = datetime.utcnow() + timedelta(minutes=10)

            # Close the temp file, this will delete the file
            temp_file.close()

        return {
            "statusCode": 200,
            "path": blob.generate_signed_url(expiration_time),
            "msg": "Successfully uploaded file to GCS"
        }

    except Exception as err:
        return {
            "statusCode": 500,
            "path": "",
            "msg": "Error Uploading file to GCS: " + str(err)
        }
