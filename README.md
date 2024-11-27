# Kambing Cup Backend

This repository contains the backend for the Kambing Cup livescore website, which provides real-time updates and scores for the Kambing Cup tournament. The backend has been migrated from Firebase to a Golang REST API to improve scalability and performance.

## Table of Contents

- [Kambing Cup Backend](#kambing-cup-backend)
  - [Table of Contents](#table-of-contents)
  - [Project Overview](#project-overview)
  - [Tech Stack](#tech-stack)
  - [Installation](#installation)

## Project Overview

The Kambing Cup backend powers the livescore website, offering real-time updates of the Kambing Cup tournament scores and stats. The goal of this revamp is to switch from Firebase to a more flexible and scalable solution using a Golang-based REST API.

Live website: [Kambing Cup Livescore](https://kambing-cup-livescore-v2.vercel.app/)

## Tech Stack

- **Language:** Golang
- **Database:** PostgreSQL 
- **Framework:** Chi
- **Database Driver:** pgx

## Installation

1. Clone the repository:

```bash
git clone https://github.com/fikrialwan/kambing-cup-backend.git
```

2. Navigate to the project directory:

```bash
cd kambing-cup-backend
```

3. Install dependencies:

```bash
go mod tidy
```

4. Start the server:

```bash
go run main.go
```
