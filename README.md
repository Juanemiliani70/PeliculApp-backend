# PeliculApp – Backend API (Golang + Gin + MongoDB)

# PeliculApp is a RESTful backend built with Golang, Gin, and MongoDB.
# It provides user authentication, movie and genre management, admin reviews,
# and secure JWT token handling (Access + Refresh tokens).

##########################
# Tech Stack
##########################
# Go 1.22+
# Gin Gonic
# MongoDB
# JWT (Access + Refresh)
# Gin Middleware
# CORS
# Go Mongo Driver v2

##########################
# Project Structure
##########################
# PeliculAppServer/
# ├── controllers/       # API logic for movies, genres, and users
# ├── database/          # MongoDB connection
# ├── middleware/        # JWT authentication
# ├── models/            # Data models: User, Movie, Genre
# ├── routes/            # Protected & public routes
# ├── utils/             # Token generation & validation
# ├── main.go            # Entry point
# ├── go.mod             # Go module file
# └── .env               # Environment variables

##########################
# Installation
##########################

# 1- Clone the repository
git clone https://github.com/Juanemiliani70/PeliculApp.git
cd PeliculApp/PeliculAppServer

# 2- Install dependencies
go mod tidy

# 3- Create your `.env` file
# Example:
cat <<EOT >> .env
# MongoDB
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=movies-app

# Server
PORT=8080
ALLOWED_ORIGINS=http://localhost:5173

# JWT Keys
SECRET_KEY=your_access_token_secret
SECRET_REFRESH_KEY=your_refresh_token_secret
EOT

##########################
# Database
##########################
# MongoDB collections:
# - users
# - movies
# - genres
# Database connection is handled in: database/connect.go

##########################
# Authentication
##########################
# JWT Access Token (24h)
# JWT Refresh Token (7 days)
# Tokens stored in MongoDB
# Access token sent as HTTPOnly cookie

# Functions:
# - Generate tokens → utils.GenerateAllTokens()
# - Validate tokens → utils.ValidateToken()
# - Refresh tokens → /refresh

##########################
# Public Routes
##########################
# Method | Route                | Description
# -------|---------------------|------------------------------
# GET    | /movies             | Get all movies
# GET    | /movie/:imdb_id     | Get a movie by IMDb ID
# GET    | /genres             | Get all genres
# GET    | /search?query=      | Search movies by title/genre
# POST   | /register           | Register a new user
# POST   | /login              | Login user
# POST   | /logout             | Logout user
# POST   | /refresh            | Refresh access token

##########################
# Protected Routes (JWT Required)
##########################
# Method | Route                   | Description
# -------|------------------------|-----------------------------------
# POST   | /addmovie               | Add a new movie (ADMIN only)
# PATCH  | /updatereview/:imdb_id  | Update admin review (ADMIN only)

##########################
# Models
##########################

# Movie (models.Movie)
# ID          ObjectID
# ImdbID      string
# Title       string
# PosterPath  string
# YouTubeID   string
# Genre       []Genre
# AdminReview string
# Description string
# WatchURL    string

# User (models.User)
# ID              ObjectID
# UserID          string
# FirstName       string
# LastName        string
# Email           string
# Password        string
# Role            string // ADMIN or USER
# FavouriteGenres []Genre
# Token           string
# RefreshToken    string
# CreatedAt       time.Time
# UpdatedAt       time.Time

# Genre (models.Genre)
# GenreID   int
# GenreName string

##########################
# Middleware
##########################
# AuthMiddleware validates JWT, extracts userId and role, 
# blocks unauthorized access (required for admin routes).

##########################
# Running the Server
##########################
go run main.go

# Server will run on:
# http://localhost:8080

##########################
# Notes
##########################
# CORS configured via ALLOWED_ORIGINS in .env
# Admin-only actions: /addmovie, /updatereview/:imdb_id
# Tokens are automatically updated in MongoDB on login and refresh
# Cookies are set with HttpOnly for security




