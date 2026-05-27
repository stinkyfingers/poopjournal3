# Plan: ReactJS UI for PoopJournal API

## Overview
This plan describes how to build a ReactJS frontend for the PoopJournal API in a separate repository. The UI will use Auth0 for authentication, provide CRUD operations for foods and poops, and be deployed as a static site on S3 + CloudFront. Terraform will be used for infrastructure.

---

## 1. Project Structure
- / (root)
  - /src (React app source)
  - /public (static assets)
  - /terraform (infra as code)
  - README.md
  - .gitignore

## 2. API Endpoints & Documentation

### Authentication
- **GET /login-page**: Returns auth status and login path.
- **GET /login**: Redirects to Auth0 login.
- **GET /callback**: Auth0 callback handler.
- **POST /logout**: Logs out the user.

### Home/Dashboard
- **GET /**: Returns dashboard summary (user, recent foods, recent poops).

### Food
- **GET /food**: List all foods and tags for the user.
- **POST /food**: Add a new food. Body: `{ name, description, datetime, tags }`
- **PUT /food/{id}**: Update a food. Body: `{ name, description, datetime, tags }`
- **DELETE /food/{id}**: Delete a food.

### Poop
- **GET /poop**: List all poops and tags for the user.
- **POST /poop**: Add a new poop. Body: `{ bristol_scale, urgency, notes, datetime, tags }`
- **PUT /poop/{id}**: Update a poop. Body: `{ bristol_scale, urgency, notes, datetime, tags }`
- **DELETE /poop/{id}**: Delete a poop.

---

## 3. UI Pages & Features

### Home Page
- Shows summary of recent foods and poops (from `/` endpoint)
- Links to Food and Poop pages

### Food Page
- List all foods
- Add/Edit/Delete food entries
- Show tags

### Poop Page
- List all poops
- Add/Edit/Delete poop entries
- Show tags

### Auth
- Use Auth0 for login/logout
- Protect all pages except login

---

## 4. Infrastructure (Terraform)
- S3 bucket for static site hosting
- CloudFront distribution for CDN/HTTPS
- Route53 DNS (optional, if you want a custom domain)
- Outputs for site URL

---

## 5. Steps to Build
1. Scaffold React app (e.g., with Create React App or Vite)
2. Implement Auth0 integration (SPA SDK)
3. Build API client for endpoints above
4. Create Home, Food, Poop pages with CRUD
5. Add Terraform for S3/CloudFront infra
6. Set up CI/CD (optional)

---

## 6. Notes
- The API base URL should be configurable (env or config file)
- Use HTTPS for all API calls
- Keep API token in memory (not localStorage) for security
- Document all environment variables needed for Auth0 and API

---

## 7. References
- [Auth0 React Quickstart](https://auth0.com/docs/quickstart/spa/react)
- [AWS S3 Static Site](https://docs.aws.amazon.com/AmazonS3/latest/userguide/WebsiteHosting.html)
- [AWS CloudFront](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Introduction.html)
- [Terraform AWS S3 Website Example](https://registry.terraform.io/modules/terraform-aws-modules/s3-bucket/aws/latest)

---

## 8. Next Steps
- Copy this plan to your new UI repo
- Start with React app scaffolding and Auth0 setup
- Use the API endpoint docs above for integration
- Use Terraform to deploy the static site
