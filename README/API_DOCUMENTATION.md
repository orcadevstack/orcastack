# orcastack - API Documentation

## Base URL
- Development: `http://localhost:8000/api`
- Production: `https://orcastack.org/api`

## Authentication

All admin endpoints require JWT authentication. Include the token in the Authorization header:
```
Authorization: Bearer <access_token>
```

### Obtain Token
```http
POST /token/
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

**Response:**
```json
{
  "access": "eyJ0eXAiOiJKV1QiLCJhbGc...",
  "refresh": "eyJ0eXAiOiJKV1QiLCJhbGc..."
}
```

### Refresh Token
```http
POST /token/refresh/
Content-Type: application/json

{
  "refresh": "eyJ0eXAiOiJKV1QiLCJhbGc..."
}
```

## Profile Endpoints

### Get Public Profile
```http
GET /accounts/profiles/public/
```

**Response:**
```json
{
  "id": 1,
  "full_name": "John Doe",
  "title": "Full-Stack Developer",
  "bio": "Brief introduction...",
  "about": "Detailed background...",
  "avatar": "http://localhost:8000/media/avatars/profile.jpg",
  "email": "john@example.com",
  "phone": "+1234567890",
  "location": "New York, USA",
  "linkedin_url": "https://linkedin.com/in/johndoe",
  "github_url": "https://github.com/johndoe",
  "skills": ["React", "Django", "Python", "JavaScript"],
  "is_active": true
}
```

### Update Profile (Admin Only)
```http
PUT /accounts/profiles/{id}/
Authorization: Bearer <token>
Content-Type: application/json

{
  "full_name": "John Doe",
  "title": "Senior Full-Stack Developer",
  "bio": "Updated bio...",
  ...
}
```

## Project Endpoints

### List All Projects
```http
GET /portfolio/projects/
GET /portfolio/projects/?category=web
GET /portfolio/projects/?featured=true
```

**Response:**
```json
[
  {
    "id": 1,
    "title": "E-commerce Platform",
    "description": "Full-featured online store",
    "detailed_description": "Detailed project description...",
    "category": "web",
    "technologies": ["React", "Django", "PostgreSQL"],
    "live_url": "https://example.com",
    "github_url": "https://github.com/user/project",
    "thumbnail": "http://localhost:8000/media/projects/thumb.jpg",
    "client": "ABC Corp",
    "completion_date": "2023-12-01",
    "is_featured": true,
    "is_published": true,
    "order": 1
  }
]
```

### Get Project Detail
```http
GET /portfolio/projects/{id}/
```

### Get Featured Projects
```http
GET /portfolio/projects/featured/
```

### Create Project (Admin Only)
```http
POST /portfolio/projects/
Authorization: Bearer <token>
Content-Type: multipart/form-data

{
  "title": "New Project",
  "description": "Project description",
  "category": "web",
  "technologies": ["React", "Node.js"],
  "thumbnail": <file>,
  "is_published": true
}
```

### Update Project (Admin Only)
```http
PUT /portfolio/projects/{id}/
Authorization: Bearer <token>
```

### Delete Project (Admin Only)
```http
DELETE /portfolio/projects/{id}/
Authorization: Bearer <token>
```

## Service Endpoints

### List All Services
```http
GET /services/
```

**Response:**
```json
[
  {
    "id": 1,
    "title": "Web Development",
    "description": "Custom web applications",
    "icon": "💻",
    "features": [
      "Responsive design",
      "SEO optimization",
      "Performance tuning"
    ],
    "pricing": "Starting at $5,000",
    "is_active": true,
    "order": 1
  }
]
```

### Create Service (Admin Only)
```http
POST /services/
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Consulting",
  "description": "Technical consulting services",
  "features": ["Architecture design", "Code review"],
  "pricing": "$200/hour",
  "is_active": true
}
```

### Update Service (Admin Only)
```http
PUT /services/{id}/
Authorization: Bearer <token>
```

### Delete Service (Admin Only)
```http
DELETE /services/{id}/
Authorization: Bearer <token>
```

## Testimonial Endpoints

### List All Testimonials
```http
GET /testimonials/
GET /testimonials/?featured=true
```

**Response:**
```json
[
  {
    "id": 1,
    "client_name": "Jane Smith",
    "client_title": "CEO",
    "client_company": "Tech Corp",
    "client_avatar": "http://localhost:8000/media/testimonials/jane.jpg",
    "content": "Excellent work! Highly recommended.",
    "rating": 5,
    "project": "E-commerce Platform",
    "is_featured": true,
    "is_published": true
  }
]
```

### Get Featured Testimonials
```http
GET /testimonials/featured/
```

### Create Testimonial (Admin Only)
```http
POST /testimonials/
Authorization: Bearer <token>
Content-Type: application/json

{
  "client_name": "John Client",
  "client_title": "CTO at StartupXYZ",
  "content": "Amazing developer!",
  "rating": 5,
  "is_published": true
}
```

### Update Testimonial (Admin Only)
```http
PUT /testimonials/{id}/
Authorization: Bearer <token>
```

### Delete Testimonial (Admin Only)
```http
DELETE /testimonials/{id}/
Authorization: Bearer <token>
```

## Inquiry/Contact Endpoints

### Submit Inquiry (Public)
```http
POST /contact/inquiries/
Content-Type: application/json

{
  "name": "Potential Client",
  "email": "client@example.com",
  "phone": "+1234567890",
  "company": "Client Corp",
  "inquiry_type": "project",
  "subject": "New Website Project",
  "message": "I need a website...",
  "budget": "$10,000 - $20,000"
}
```

**Response:**
```json
{
  "message": "Your inquiry has been submitted successfully!",
  "id": 5
}
```

### List All Inquiries (Admin Only)
```http
GET /contact/inquiries/
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": 1,
    "name": "Potential Client",
    "email": "client@example.com",
    "phone": "+1234567890",
    "company": "Client Corp",
    "inquiry_type": "project",
    "subject": "New Website Project",
    "message": "I need a website...",
    "budget": "$10,000 - $20,000",
    "status": "new",
    "created_at": "2023-12-06T10:30:00Z"
  }
]
```

### Update Inquiry Status (Admin Only)
```http
PATCH /contact/inquiries/{id}/update_status/
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "completed"
}
```

**Status Options:**
- `new` - New inquiry
- `in_progress` - Being worked on
- `completed` - Completed
- `archived` - Archived

## Field Types

### Project Categories
- `web` - Web Development
- `mobile` - Mobile Development
- `design` - Design
- `consulting` - Consulting
- `other` - Other

### Inquiry Types
- `general` - General Inquiry
- `project` - Project Request
- `job` - Job Opportunity
- `collaboration` - Collaboration
- `other` - Other

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid data provided",
  "details": {
    "email": ["Enter a valid email address."]
  }
}
```

### 401 Unauthorized
```json
{
  "detail": "Authentication credentials were not provided."
}
```

### 403 Forbidden
```json
{
  "detail": "You do not have permission to perform this action."
}
```

### 404 Not Found
```json
{
  "error": "Not found"
}
```

## Rate Limiting

Currently no rate limiting is implemented. Consider adding in production using Django REST Framework throttling.

## Pagination

List endpoints return paginated results:
```json
{
  "count": 25,
  "next": "http://localhost:8000/api/portfolio/projects/?page=2",
  "previous": null,
  "results": [...]
}
```

Use `?page=N` to navigate pages.

## File Uploads

For endpoints that accept files (thumbnails, avatars):
- Use `multipart/form-data` content type
- Supported formats: JPG, PNG, GIF
- Max file size: 5MB (configurable in settings)

Example with curl:
```bash
curl -X POST http://localhost:8000/api/portfolio/projects/ \
  -H "Authorization: Bearer <token>" \
  -F "title=New Project" \
  -F "description=Description here" \
  -F "category=web" \
  -F "thumbnail=@/path/to/image.jpg"
```

## CORS

The API supports CORS for configured origins. In development:
- `http://localhost:3000`
- `http://127.0.0.1:3000`

Update `CORS_ALLOWED_ORIGINS` in production settings.
