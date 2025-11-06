# 🐋 Docker Testing Guide - Step by Step

## 📋 Prerequisites

Before starting, make sure you have:
- ✅ Docker Desktop installed and running
- ✅ Internet connection (for pulling images)
- ✅ Port 8080, 27017, 6379 available on your machine

---

## 🚀 Step 1: Start All Services with Docker

### Option A: Start Without Management Tools (Recommended for Testing)

```powershell
# Navigate to project directory
cd "c:\Users\hp\OneDrive\Documents\GitHub\Lujay assesment"

# Build and start all services
docker-compose up --build -d

# Check if all services are running
docker-compose ps
```

**Expected Output:**
```
NAME                COMMAND                  SERVICE    STATUS         PORTS
lujay-app           "/app/server"            app        Up (healthy)   0.0.0.0:8080->8080/tcp
lujay-mongodb       "docker-entrypoint.s…"   mongodb    Up (healthy)   0.0.0.0:27017->27017/tcp
lujay-redis         "docker-entrypoint.s…"   redis      Up (healthy)   0.0.0.0:6379->6379/tcp
```

### Option B: Start WITH Management Tools (MongoDB & Redis UI)

```powershell
# Start all services including management tools
docker-compose --profile tools up --build -d

# Check if all services are running
docker-compose ps
```

**This gives you access to:**
- 🌐 **API**: http://localhost:8080
- 🍃 **Mongo Express**: http://localhost:8082 (Username: `admin`, Password: `admin123`)
- 🔴 **Redis Commander**: http://localhost:8081

---

## 🔍 Step 2: Verify Services are Running

### Check Logs

```powershell
# View all logs
docker-compose logs

# View only app logs
docker-compose logs app

# Follow app logs in real-time
docker-compose logs -f app

# View last 50 lines
docker-compose logs --tail=50 app
```

**What to look for in logs:**
```
✅ Server started successfully on :8080
✅ Connected to MongoDB successfully
✅ Redis connection established
✅ Cloudinary uploader initialized
```

### Check Health Status

```powershell
# Check health of all containers
docker-compose ps

# Check specific service health
docker inspect --format='{{.State.Health.Status}}' lujay-app
docker inspect --format='{{.State.Health.Status}}' lujay-mongodb
docker inspect --format='{{.State.Health.Status}}' lujay-redis
```

---

## 🧪 Step 3: Test the API

### Test 1: Health Check

```powershell
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-06T..."
}
```

### Test 2: Register a New User

```powershell
# Register as a Dealer (can create vehicles)
curl -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{
    "name": "Test Dealer",
    "email": "dealer@test.com",
    "password": "Password123!",
    "phoneNumber": "+2348012345678",
    "role": "dealer"
  }'
```

**Expected Response:**
```json
{
  "message": "user registered successfully",
  "user": {
    "id": "...",
    "name": "Test Dealer",
    "email": "dealer@test.com",
    "role": "dealer"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**💡 IMPORTANT: Copy the token from the response!**

### Test 3: Login

```powershell
curl -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{
    "email": "dealer@test.com",
    "password": "Password123!"
  }'
```

### Test 4: Create a Vehicle

```powershell
# Replace YOUR_TOKEN_HERE with your actual token
$token = "YOUR_TOKEN_HERE"

curl -X POST http://localhost:8080/api/v1/vehicles `
  -H "Authorization: Bearer $token" `
  -H "Content-Type: application/json" `
  -d '{
    "make": "Toyota",
    "model": "Camry",
    "year": 2024,
    "price": 45000,
    "mileage": 0,
    "condition": "new",
    "fuelType": "petrol",
    "transmission": "automatic",
    "bodyType": "sedan",
    "exteriorColor": "black",
    "interiorColor": "beige",
    "numberOfDoors": 4,
    "numberOfSeats": 5,
    "vin": "1HGBH41JXMN109186",
    "location": {
      "address": "123 Car Street, Lagos",
      "city": "Lagos",
      "state": "Lagos",
      "country": "Nigeria"
    }
  }'
```

**Expected Response:**
```json
{
  "message": "vehicle created successfully",
  "vehicle": {
    "id": "672b...",
    "make": "Toyota",
    "model": "Camry",
    ...
  }
}
```

**💡 IMPORTANT: Copy the vehicle ID from the response!**

---

## 📸 Step 4: Test File Upload (THE MAIN FEATURE!)

### Using PowerShell with curl

```powershell
# Set your token and vehicle ID
$token = "YOUR_TOKEN_HERE"
$vehicleId = "YOUR_VEHICLE_ID_HERE"

# Upload images (replace with actual file paths)
curl -X POST "http://localhost:8080/api/v1/vehicles/$vehicleId/images" `
  -H "Authorization: Bearer $token" `
  -F "images=@C:\Users\hp\Pictures\car1.jpg" `
  -F "images=@C:\Users\hp\Pictures\car2.jpg" `
  -F "images=@C:\Users\hp\Pictures\car3.jpg"
```

**Expected Response:**
```json
{
  "message": "images uploaded successfully",
  "images_added": 3,
  "total_images": 3,
  "uploaded_images": [
    {
      "url": "https://res.cloudinary.com/dksiuqlwq/image/upload/v.../lujay/vehicles/vehicle_672b.../image_1730...",
      "public_id": "lujay/vehicles/vehicle_672b.../image_1730...",
      "format": "jpg",
      "width": 1920,
      "height": 1080,
      "bytes": 245632,
      "isPrimary": true
    },
    ...
  ]
}
```

### Using Postman (EASIER!)

1. **Open Postman** (Download from https://www.postman.com/downloads/)

2. **Create New Request:**
   - Method: `POST`
   - URL: `http://localhost:8080/api/v1/vehicles/{vehicleId}/images`
   - Replace `{vehicleId}` with your actual vehicle ID

3. **Set Headers:**
   - Click "Headers" tab
   - Add: `Authorization` = `Bearer YOUR_TOKEN_HERE`

4. **Set Body:**
   - Click "Body" tab
   - Select "form-data"
   - Add key: `images`, Type: `File` (change from "Text" to "File")
   - Click "Select Files" and choose your images
   - You can add multiple rows with key `images` for multiple files

5. **Click Send**

### Test Delete Image

```powershell
# Get the public_id from upload response
$publicId = "lujay/vehicles/vehicle_672b.../image_1730..."

curl -X DELETE "http://localhost:8080/api/v1/vehicles/$vehicleId/images/$publicId" `
  -H "Authorization: Bearer $token"
```

### Test Set Primary Image

```powershell
curl -X PUT "http://localhost:8080/api/v1/vehicles/$vehicleId/images/$publicId/primary" `
  -H "Authorization: Bearer $token"
```

---

## 🔍 Step 5: Verify in Cloudinary Dashboard

1. Go to https://cloudinary.com and login
2. Click "Media Library" in the left sidebar
3. Navigate to `lujay/vehicles/` folder
4. You should see your uploaded images organized by vehicle ID!

---

## 🎯 Step 6: View Data in Database (Optional)

### Using Mongo Express (if you started with `--profile tools`)

1. Open http://localhost:8082
2. Login: `admin` / `admin123`
3. Click `lujay_db` database
4. View collections: `vehicles`, `users`, `inspections`, `transactions`
5. Click any collection to see documents

### Using Redis Commander (if you started with `--profile tools`)

1. Open http://localhost:8081
2. Browse cached data
3. See rate limiting counters
4. Monitor cache hits/misses

---

## 🛠️ Useful Docker Commands

### View Logs

```powershell
# All logs
docker-compose logs

# Specific service logs
docker-compose logs app
docker-compose logs mongodb
docker-compose logs redis

# Follow logs (real-time)
docker-compose logs -f app

# Last 100 lines
docker-compose logs --tail=100 app
```

### Restart Services

```powershell
# Restart all services
docker-compose restart

# Restart specific service
docker-compose restart app
```

### Stop Services

```powershell
# Stop all services (keeps data)
docker-compose stop

# Stop and remove containers (keeps volumes/data)
docker-compose down

# Stop and remove everything including volumes (DELETES DATA!)
docker-compose down -v
```

### Rebuild After Code Changes

```powershell
# Rebuild and restart
docker-compose up --build -d

# Rebuild specific service
docker-compose up --build -d app
```

### Access Container Shell

```powershell
# Access app container
docker exec -it lujay-app sh

# Access MongoDB container
docker exec -it lujay-mongodb mongosh lujay_db

# Access Redis container
docker exec -it lujay-redis redis-cli
```

### Check Resource Usage

```powershell
# See CPU, memory, network usage
docker stats

# See disk usage
docker system df
```

---

## 🐛 Troubleshooting

### Problem: Port Already in Use

**Error:**
```
Bind for 0.0.0.0:8080 failed: port is already allocated
```

**Solution:**
```powershell
# Find process using port 8080
netstat -ano | findstr :8080

# Kill the process (replace PID with actual process ID)
taskkill /PID <PID> /F

# Or change port in docker-compose.yml
# Change: "8080:8080" to "8081:8080"
```

### Problem: Container Unhealthy

**Error:**
```
lujay-app    Up (unhealthy)
```

**Solution:**
```powershell
# Check logs for errors
docker-compose logs app

# Common issues:
# 1. MongoDB not ready - wait 30 seconds and check again
# 2. Missing .env variables - verify .env file
# 3. Cloudinary credentials invalid - check CLOUDINARY_* values
```

### Problem: Docker Desktop Not Running

**Error:**
```
Cannot connect to the Docker daemon
```

**Solution:**
1. Open Docker Desktop application
2. Wait for it to start (whale icon in system tray should be running)
3. Try command again

### Problem: Images Not Uploading

**Check these:**

```powershell
# 1. Check app logs
docker-compose logs app | Select-String -Pattern "cloudinary"

# 2. Verify Cloudinary credentials in .env
# Should NOT be empty:
# CLOUDINARY_CLOUD_NAME=dksiuqlwq
# CLOUDINARY_API_KEY=516115242333556
# CLOUDINARY_API_SECRET=F7KWWqkhLAybpNkoUWALiRBc6Q4

# 3. Restart app after .env changes
docker-compose restart app

# 4. Check file size (max 10MB per file)
# 5. Check file format (jpg, jpeg, png, gif, webp only)
```

### Problem: Can't Access Database from Host

**Solution:**
```powershell
# MongoDB connection string for external tools (like MongoDB Compass):
mongodb://localhost:27017/lujay_db

# Redis connection for external tools:
# Host: localhost
# Port: 6379
# Password: (empty)
```

---

## 📊 Complete Testing Workflow

### Full End-to-End Test Script

```powershell
# 1. Start services
docker-compose up --build -d

# 2. Wait for services to be healthy (30 seconds)
timeout 30

# 3. Check health
curl http://localhost:8080/health

# 4. Register user
$registerResponse = curl -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{
    "name": "Test Dealer",
    "email": "dealer@test.com",
    "password": "Password123!",
    "phoneNumber": "+2348012345678",
    "role": "dealer"
  }' | ConvertFrom-Json

$token = $registerResponse.token
Write-Host "Token: $token"

# 5. Create vehicle
$vehicleResponse = curl -X POST http://localhost:8080/api/v1/vehicles `
  -H "Authorization: Bearer $token" `
  -H "Content-Type: application/json" `
  -d '{
    "make": "Toyota",
    "model": "Camry",
    "year": 2024,
    "price": 45000,
    "mileage": 0,
    "condition": "new",
    "fuelType": "petrol",
    "transmission": "automatic",
    "bodyType": "sedan",
    "exteriorColor": "black",
    "interiorColor": "beige",
    "numberOfDoors": 4,
    "numberOfSeats": 5,
    "vin": "1HGBH41JXMN109186",
    "location": {
      "address": "123 Car Street",
      "city": "Lagos",
      "state": "Lagos",
      "country": "Nigeria"
    }
  }' | ConvertFrom-Json

$vehicleId = $vehicleResponse.vehicle.id
Write-Host "Vehicle ID: $vehicleId"

# 6. Upload images (replace with your image paths)
curl -X POST "http://localhost:8080/api/v1/vehicles/$vehicleId/images" `
  -H "Authorization: Bearer $token" `
  -F "images=@C:\path\to\your\image.jpg"

# 7. Get vehicle with images
curl -X GET "http://localhost:8080/api/v1/vehicles/$vehicleId" `
  -H "Authorization: Bearer $token"

Write-Host "✅ All tests completed!"
```

---

## 🎯 Quick Reference

| Service | URL | Purpose |
|---------|-----|---------|
| API | http://localhost:8080 | Main application |
| Health Check | http://localhost:8080/health | Check if API is running |
| Mongo Express | http://localhost:8082 | MongoDB web UI |
| Redis Commander | http://localhost:8081 | Redis web UI |
| MongoDB Direct | mongodb://localhost:27017 | Database connection |
| Redis Direct | localhost:6379 | Cache connection |

| Command | Purpose |
|---------|---------|
| `docker-compose up -d` | Start all services |
| `docker-compose up --profile tools -d` | Start with management tools |
| `docker-compose ps` | Check service status |
| `docker-compose logs -f app` | View app logs |
| `docker-compose restart app` | Restart app |
| `docker-compose down` | Stop all services |
| `docker-compose down -v` | Stop and delete data |

---

## 🎉 Success Checklist

After following this guide, you should be able to:

- ✅ Start all services with Docker
- ✅ Register a new user
- ✅ Login and get JWT token
- ✅ Create a vehicle
- ✅ Upload images to vehicle
- ✅ See images in Cloudinary dashboard
- ✅ Delete images
- ✅ Set primary image
- ✅ View data in MongoDB
- ✅ Monitor Redis cache

---

## 💡 Pro Tips

1. **Use Postman for easier testing** - Much better than curl for file uploads
2. **Keep Docker Desktop running** - Required for all Docker commands
3. **Check logs if something fails** - `docker-compose logs app`
4. **Use management tools** - Start with `--profile tools` for debugging
5. **Save your tokens** - Keep them in a text file for testing
6. **Test incrementally** - Don't skip steps, test each endpoint
7. **Watch Cloudinary dashboard** - Verify images are actually uploaded

---

## 🚨 Common Mistakes to Avoid

1. ❌ Forgetting to include `Bearer` in Authorization header
2. ❌ Using wrong vehicle ID (copy from create vehicle response)
3. ❌ Using wrong public_id for delete (copy from upload response)
4. ❌ Not starting Docker Desktop before running commands
5. ❌ Using local MongoDB/Redis URLs instead of Docker service names
6. ❌ Uploading files larger than 10MB
7. ❌ Not waiting for services to be healthy before testing

---

## 📞 Need Help?

If you encounter issues:

1. Check logs: `docker-compose logs app`
2. Verify services are healthy: `docker-compose ps`
3. Check Cloudinary credentials in `.env`
4. Make sure ports 8080, 27017, 6379 are not in use
5. Restart services: `docker-compose restart`
6. Check Docker Desktop is running

---

**Happy Testing! 🎊**
