## Presigned Multipart Go Backend Service

This repository provides a Go backend service for initiating multipart uploads, generating presigned URLs for part uploads, and completing multipart uploads to Amazon S3.

**Features:**

- Implements a multi-part upload workflow for large files.
- Provides API endpoints for initiating, pre-signing individual parts, and completing uploads.
- Uses AWS S3 for storage.

**Requirements:**

- Go 1.17 or later
- AWS account with S3 bucket

**Installation:**

1. Clone this repository.
2. Create a `.env` file in the project root directory with the following environment variables:
   - `AWS_ACCESS_KEY_ID`: The access id with s3 permissions.
   - `AWS_SECRET_ACCESS_KEY`: The secret access key with s3 permissions.
   - `AWS_REGION`: The AWS region where your S3 bucket is located.
   - `AWS_BUCKET`: The name of your S3 bucket.
   - `PORT`: Server running port

**Running the Service:**

1. Build the service:
   ```bash
   make build
   ```
2. Run the service:
   ```bash
   ./bin/service
   ```

This will start the service and listen for requests on port 8080.

**API Endpoints:**

- **`/initiate` (GET):** Initiates a multipart upload for a file.

  - Query parameter: `key` (string): The name of the file to upload.
  - Response: JSON object containing the `uploadId` for the multipart upload.

- **`/presigned` (GET):** Generates a presigned URL for uploading a specific part of a multipart upload.

  - Query parameters:
    - `key` (string): The name of the file to upload.
    - `uploadId` (string): The `uploadId` obtained from the `/initiate` endpoint.
    - `partNumber` (int): The part number of the part to upload.
  - Response: JSON object containing the `uploadUrl` for the part upload, the `uploadId`, the `partNumber`, and the `expiresAt` timestamp for the presigned URL.

- **`/complete` (POST):** Completes a multipart upload by providing the ETags of all uploaded parts.
  - Request body: JSON object with a `parts` array containing objects with `PartNumber` (int) and `ETag` (string) keys for each uploaded part.
  - Query parameters:
    - `key` (string): The name of the file that was uploaded.
    - `uploadId` (string): The `uploadId` obtained from the `/initiate` endpoint.
  - Response: Status code 200 if the upload is completed successfully.

**CORS Configuration:**

The service is configured with CORS (Cross-Origin Resource Sharing) to allow requests from any origin. This configuration might need to be adjusted depending on your specific deployment environment.

**Security Considerations:**

- This example uses environment variables for sensitive information like AWS credentials. Consider using a more secure method for storing and managing secrets in production environments.
- The CORS configuration allows requests from any origin. This configuration should be tightened for production environments.

**Contributing:**

We welcome contributions to this project! Please see the CONTRIBUTING.md file for guidelines on how to contribute.

**License:**

This project is licensed under the MIT License. See the LICENSE file for details.
