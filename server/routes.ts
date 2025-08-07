import type { Express } from "express";
import { createServer, type Server } from "http";
import { storage } from "./storage";

export async function registerRoutes(app: Express): Promise<Server> {
  // User management routes
  app.get("/api/users", async (req, res) => {
    try {
      // In a real app, you'd implement pagination and filtering
      res.json({ users: [], message: "Users endpoint working" });
    } catch (error) {
      res.status(500).json({ error: "Failed to fetch users" });
    }
  });

  app.post("/api/users", async (req, res) => {
    try {
      const userData = req.body;
      
      // Input validation and sanitization
      if (!userData.username || !userData.password) {
        return res.status(400).json({ error: "Username and password required" });
      }
      
      // Sanitize username (remove HTML/script tags)
      const username = userData.username.toString().replace(/<[^>]*>/g, '').trim();
      const password = userData.password.toString().trim();
      
      // Validate username format (alphanumeric, underscore, dash only)
      if (!/^[a-zA-Z0-9_-]{3,20}$/.test(username)) {
        return res.status(400).json({ error: "Username must be 3-20 characters, alphanumeric only" });
      }
      
      // Validate password length
      if (password.length < 8) {
        return res.status(400).json({ error: "Password must be at least 8 characters" });
      }
      
      // Only pass validated fields to storage
      const sanitizedUserData = { username, password };
      const user = await storage.createUser(sanitizedUserData);
      res.json({ user: { id: user.id, username: user.username }, message: "User created successfully" });
    } catch (error) {
      console.error("User creation error:", error);
      res.status(500).json({ error: "Failed to create user" });
    }
  });

  app.get("/api/users/:id", async (req, res) => {
    try {
      const userId = parseInt(req.params.id);
      if (isNaN(userId)) {
        return res.status(400).json({ error: "Invalid user ID" });
      }
      
      const user = await storage.getUser(userId);
      if (!user) {
        return res.status(404).json({ error: "User not found" });
      }
      
      res.json({ user: { id: user.id, username: user.username } });
    } catch (error) {
      res.status(500).json({ error: "Failed to fetch user" });
    }
  });

  // Health check endpoint
  app.get("/api/health", (req, res) => {
    res.json({ status: "ok", timestamp: new Date().toISOString() });
  });

  const httpServer = createServer(app);

  return httpServer;
}
