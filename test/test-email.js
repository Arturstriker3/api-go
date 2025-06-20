const net = require("net");
const fs = require("fs").promises;
const path = require("path");
const readline = require("readline");

class EmailTestClient {
  constructor() {
    this.config = null;
    this.emailTemplate = null;
    this.client = null;
    this.authenticated = false;
    this.rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
    });
  }

  async initialize() {
    try {
      const configPath = path.join(__dirname, "test-config.json");
      const templatePath = path.join(__dirname, "email-template.json");

      const [configData, templateData] = await Promise.all([
        fs.readFile(configPath, "utf8"),
        fs.readFile(templatePath, "utf8"),
      ]);

      this.config = JSON.parse(configData);
      this.emailTemplate = JSON.parse(templateData);
      console.log("Config loaded:", {
        host: this.config.connection.host,
        port: this.config.connection.port,
        authSecret: this.config.connection.authSecret,
      });
    } catch (error) {
      console.error("Error loading configuration:", error.message);
      process.exit(1);
    }
  }

  connect() {
    return new Promise((resolve, reject) => {
      this.client = new net.Socket();
      this.authenticated = false;

      this.client.connect(
        this.config.connection.port,
        this.config.connection.host,
        () => {
          console.log("Connected to server");
          resolve();
        }
      );

      this.client.on("error", (error) => {
        console.error("Connection error:", error.message);
        reject(error);
      });

      this.client.on("close", () => {
        console.log("Connection closed by server");
        this.authenticated = false;
      });
    });
  }

  authenticate() {
    return new Promise((resolve, reject) => {
      if (this.authenticated) {
        resolve();
        return;
      }

      const authData = {
        secret: this.config.connection.authSecret,
      };

      console.log("Sending auth data:", authData);

      const responseHandler = (data) => {
        const response = data.toString().trim();
        console.log("Auth response received:", response);
        this.client.removeListener("data", responseHandler);

        try {
          const parsed = JSON.parse(response);
          if (parsed.error) {
            reject(new Error(parsed.error));
          } else {
            this.authenticated = true;
            console.log("Authentication successful");
            resolve();
          }
        } catch (e) {
          console.error("Failed to parse auth response:", e.message);
          reject(new Error("Invalid response format"));
        }
      };

      const timeoutHandler = setTimeout(() => {
        this.client.removeListener("data", responseHandler);
        reject(new Error("Authentication timeout"));
      }, 5000);

      this.client.on("data", (data) => {
        clearTimeout(timeoutHandler);
        responseHandler(data);
      });

      this.client.write(JSON.stringify(authData) + "\n");
    });
  }

  async sendEmail(count = 1) {
    if (!this.client) {
      throw new Error("Not connected to server");
    }

    if (!this.authenticated) {
      await this.authenticate();
    }

    const results = {
      success: 0,
      failed: 0,
    };

    for (let i = 0; i < count; i++) {
      const emailData = {
        to: [this.config.email.to],
        subject: this.emailTemplate.subject,
        body: this.emailTemplate.body
          .replace("{{timestamp}}", new Date().toISOString())
          .replace("{{testNumber}}", (i + 1).toString()),
      };

      console.log(`Sending email ${i + 1}/${count}:`, {
        to: emailData.to,
        subject: emailData.subject,
      });

      try {
        await this.sendSingleEmail(emailData);
        results.success++;
        if (count > 1) {
          console.log(`Email ${i + 1}/${count} sent successfully`);
        }

        if (i < count - 1 && this.config.settings.defaultDelay) {
          console.log(
            `Waiting ${this.config.settings.defaultDelay}ms before next email...`
          );
          await new Promise((resolve) =>
            setTimeout(resolve, this.config.settings.defaultDelay)
          );
        }
      } catch (error) {
        results.failed++;
        console.error(`Failed to send email ${i + 1}:`, error.message);
      }
    }

    return results;
  }

  sendSingleEmail(emailData) {
    return new Promise((resolve, reject) => {
      const data = JSON.stringify(emailData) + "\n";
      console.log("Sending email data:", data.trim());

      const responseHandler = (data) => {
        const response = data.toString().trim();
        console.log("Email response received:", response);
        this.client.removeListener("data", responseHandler);

        try {
          const parsed = JSON.parse(response);
          if (parsed.error) {
            reject(new Error(parsed.error));
          } else {
            resolve(parsed.message || "Success");
          }
        } catch (e) {
          console.error("Failed to parse email response:", e.message);
          reject(new Error("Invalid response format"));
        }
      };

      const timeoutHandler = setTimeout(() => {
        this.client.removeListener("data", responseHandler);
        reject(new Error("Email sending timeout"));
      }, 10000);

      this.client.on("data", (data) => {
        clearTimeout(timeoutHandler);
        responseHandler(data);
      });

      this.client.write(data);
    });
  }

  async disconnect() {
    if (this.client) {
      this.client.destroy();
      console.log("Disconnected from server");
    }
    this.authenticated = false;
    this.rl.close();
  }

  question(prompt) {
    return new Promise((resolve) => {
      this.rl.question(prompt, resolve);
    });
  }

  async showMenu() {
    while (true) {
      console.log("\n1. Send single email");
      console.log("2. Send multiple emails");
      console.log("3. Exit");

      const answer = await this.question("\nChoose an option (1-3): ");

      if (answer === "3") {
        await this.disconnect();
        break;
      }

      try {
        if (!this.client) {
          await this.connect();
        }

        if (answer === "1") {
          const result = await this.sendEmail(1);
          console.log("\nResult:", result);
        } else if (answer === "2") {
          const countStr = await this.question(
            "How many emails do you want to send? "
          );
          const count = parseInt(countStr);

          if (isNaN(count) || count <= 0) {
            console.log("Please enter a valid number greater than 0");
            continue;
          }

          const result = await this.sendEmail(count);
          console.log("\nFinal Results:");
          console.log(`Successful: ${result.success}`);
          console.log(`Failed: ${result.failed}`);
        } else {
          console.log("Invalid option. Please choose 1, 2, or 3.");
        }
      } catch (error) {
        console.error("Error:", error.message);
        await this.disconnect();
      }
    }
  }
}

async function main() {
  const client = new EmailTestClient();
  await client.initialize();
  await client.showMenu();
}

main().catch(console.error);
