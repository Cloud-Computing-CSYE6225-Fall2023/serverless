#!/bin/bash

mkdir -p ./python/lib/python3.11/site-packages
pip3 install boto3 requests google-cloud-storage==2.9.0 -t ./python/lib/python3.11/site-packages

# zip python3.11 dependencies to dependencies.zip with file structure /python/lib/python3.11/site-packages
zip -r dependencies.zip ./python

# zip lambda function to gcp_file_storage.zip
zip -r gcp_file_storage.zip lambda_function.py download_zip.py gcs_create.py mail_sender.py track_email.py

# Modifiy file permissions
sudo chmod 644 dependencies.zip gcp_file_storage.zip
