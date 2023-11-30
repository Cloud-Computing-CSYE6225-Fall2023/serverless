import os
import json
import requests


def send_email(data):
    try:
        domainName = os.environ["mailgun_domain_name"]
        apiKey = os.environ["mailgun_api_key"]
        emailTemplate = os.environ["mailgun_template"]

        response = requests.post(
            url=f"https://api.mailgun.net/v3/{domainName}/messages",
            auth=("api", apiKey),
            data={"from": f"Excited User <noreply@{domainName}>",
                  "to": [data["email"]],
                  "template": emailTemplate,
                  "t:variables": json.dumps({
                      "status": data["status"],
                      "submission_url": data["submission_url"],
                      "file_path": data["path"],
                      "submission_time": data["submission_date"]
                    })
                  })
        print("here1: ", response.json())

        return {
            "statusCode": response.status_code,
            "msg": response.json()
        }

    except Exception as err:
        print("here 3: ", str(err))
        return {
            "statusCode": 500,
            "msg": "Error Sending Email " + str(err)
        }
