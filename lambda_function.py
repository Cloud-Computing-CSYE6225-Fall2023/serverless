import json
from mail_sender import send_email
from gcs_create import write_to_blob
from track_email import track_emails_to_db
from download_zip import download_and_check_zip


def lambda_handler(event, context):
    event_data = json.loads(event["Records"][0]["Sns"]["Message"])
    file_path = f'{event_data["assignment_id"]}/{event_data["user"]["id"]}/{event_data["submission_date"]}.zip'
    file_download_resp = download_and_check_zip(event_data["submission_url"])

    data = {
        "id": event_data["id"],
        "assignment_id": event_data["assignment_id"],
        "user_id":  event_data["user"]["id"],
        "email": event_data["user"]["email"],
        "submission_url": event_data["submission_url"],
        "submission_date": event_data["submission_date"]
    }

    if file_download_resp["statusCode"] == 200:
        file_upload_resp = write_to_blob(file_download_resp["data"], file_path)

        if file_upload_resp["statusCode"] == 200:
            data["status"] = "successfully downloaded!"
            data["path"] = file_upload_resp["path"]

            send_email(data)  # Send email
            track_emails_to_db(data)  # Update to dynamo DB

            return {
                "statusCode": 200,
                "msg": "File Successfully uploaded to GCS"
            }

        elif file_upload_resp["statusCode"] == 500:
            data["status"] = "failed to download, please recheck the url submitted and the file contents"
            data["path"] = ""

            send_email(data)  # Send email
            track_emails_to_db(data)  # Update to dynamo DB

            return file_upload_resp

    elif file_download_resp["statusCode"] == 500:
        data["status"] = "failed to download, please recheck the url submitted and the file contents"
        data["path"] = ""

        send_email(data)  # Send email
        track_emails_to_db(data)  # Update to dynamo DB

        return file_download_resp


# if __name__ == "__main__":
#     d = {
#         "assignment_id": "123",
#         "id": "9181",
#         "user": {
#             "id": "7121",
#             "email": "ruthala.s@northeastern.edu"
#         },
#         "submission_url": "https://github.com/tparikh/myrepo/archive/refs/tags/v1.0.0.zip",
#         "submission_date": "2023-11-24T09:00"
#     }
#     print(lambda_handler(d, ""))
