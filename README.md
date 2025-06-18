# House-Pricer (OLX Warsaw Flat Offers Scraper)

**House-Pricer** is a serverless Go application designed to fetch and parse flat rental offers from OLX (Warsaw), 
structure them into a defined format, and store them in an AWS DynamoDB table. 
It runs as an AWS Lambda function and is fully deployed using Terraform.

---

## Features

- Scrapes OLX.pl for flat rental offers in **Warsaw**
- Extracts structured information from each offer:
  - URL
  - Title
  - Price (PLN/month)
  - Area (in squared meters)
  - Estimated mortgage
  - Furnished status
  - Timestamps for creation and update
- Persists data in **AWS DynamoDB**
- Runs as an **AWS Lambda** function

---

## Purpose
This scraper is intended to become the backbone of a larger ecosystem for monitoring rental markets. T
he long-term goal is to support scraping from multiple rental websites, not just OLX.
Once multi-source support is in place, a notification system will be implemented. 
This system will allow users to define strict filters (e.g., price range, location, furnished status), and receive alerts whenever matching offers appear online.


