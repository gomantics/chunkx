const express = require("express");
const { body, validationResult } = require("express-validator");

/**
 * UserController handles all user-related HTTP endpoints
 */
class UserController {
  constructor(userService, logger) {
    this.userService = userService;
    this.logger = logger;
  }

  /**
   * Get all users with optional filtering and pagination
   */
  async getUsers(req, res) {
    try {
      const { page = 1, limit = 10, search = "" } = req.query;

      const options = {
        page: parseInt(page),
        limit: parseInt(limit),
        search: search.trim(),
      };

      const users = await this.userService.findAll(options);
      const total = await this.userService.count(options.search);

      res.json({
        data: users,
        pagination: {
          page: options.page,
          limit: options.limit,
          total: total,
          pages: Math.ceil(total / options.limit),
        },
      });
    } catch (error) {
      this.logger.error("Error fetching users:", error);
      res.status(500).json({ error: "Internal server error" });
    }
  }

  /**
   * Get a single user by ID
   */
  async getUserById(req, res) {
    try {
      const { id } = req.params;
      const user = await this.userService.findById(id);

      if (!user) {
        return res.status(404).json({ error: "User not found" });
      }

      res.json({ data: user });
    } catch (error) {
      this.logger.error(`Error fetching user ${req.params.id}:`, error);
      res.status(500).json({ error: "Internal server error" });
    }
  }

  /**
   * Create a new user
   */
  async createUser(req, res) {
    try {
      const errors = validationResult(req);
      if (!errors.isEmpty()) {
        return res.status(400).json({ errors: errors.array() });
      }

      const userData = {
        email: req.body.email,
        name: req.body.name,
        role: req.body.role || "user",
      };

      const user = await this.userService.create(userData);

      this.logger.info(`User created: ${user.id}`);
      res.status(201).json({ data: user });
    } catch (error) {
      if (error.code === "DUPLICATE_EMAIL") {
        return res.status(409).json({ error: "Email already exists" });
      }
      this.logger.error("Error creating user:", error);
      res.status(500).json({ error: "Internal server error" });
    }
  }

  /**
   * Update an existing user
   */
  async updateUser(req, res) {
    try {
      const { id } = req.params;
      const errors = validationResult(req);

      if (!errors.isEmpty()) {
        return res.status(400).json({ errors: errors.array() });
      }

      const updates = {
        email: req.body.email,
        name: req.body.name,
        role: req.body.role,
      };

      const user = await this.userService.update(id, updates);

      if (!user) {
        return res.status(404).json({ error: "User not found" });
      }

      this.logger.info(`User updated: ${id}`);
      res.json({ data: user });
    } catch (error) {
      this.logger.error(`Error updating user ${req.params.id}:`, error);
      res.status(500).json({ error: "Internal server error" });
    }
  }

  /**
   * Delete a user
   */
  async deleteUser(req, res) {
    try {
      const { id } = req.params;
      const deleted = await this.userService.delete(id);

      if (!deleted) {
        return res.status(404).json({ error: "User not found" });
      }

      this.logger.info(`User deleted: ${id}`);
      res.status(204).send();
    } catch (error) {
      this.logger.error(`Error deleting user ${req.params.id}:`, error);
      res.status(500).json({ error: "Internal server error" });
    }
  }
}

module.exports = UserController;
