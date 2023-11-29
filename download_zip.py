import requests


def download_and_check_zip(url):
    try:
        response = requests.get(url, stream=True)
        if response.status_code == 200:
            return {
                "statusCode": 200,
                "data": response
            }
        else:
            return {
                "statusCode": 500,
                "msg": "Invalid url"
            }

    except requests.exceptions.RequestException as err:
        return {
            "statusCode": 500,
            "msg": "Error downloading file from url: " + err.response,
        }
    except Exception as err:
        return {
            "statusCode": 500,
            "msg": "Error downloading file from url: " + str(err),
        }

