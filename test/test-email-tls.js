const tls = require("tls");
const fs = require("fs").promises;
const path = require("path");
const readline = require("readline");

class EmailTestTLSClient {
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

      console.log("🔒 TLS Config loaded:", {
        host: this.config.connection.host,
        port: this.config.connection.port,
        tlsEnabled: this.config.connection.tlsEnabled || false,
        rejectUnauthorized: this.config.connection.rejectUnauthorized !== false,
      });
    } catch (error) {
      console.error("❌ Error loading configuration:", error.message);
      process.exit(1);
    }
  }

  connect() {
    return new Promise((resolve, reject) => {
      this.authenticated = false;

      const options = {
        host: this.config.connection.host,
        port: this.config.connection.port,
        rejectUnauthorized: this.config.connection.rejectUnauthorized !== false,
      };

      // Add CA certificate if specified and file exists
      if (this.config.connection.caPath) {
        try {
          const fs = require("fs");
          if (fs.existsSync(this.config.connection.caPath)) {
            const ca = fs.readFileSync(this.config.connection.caPath);
            options.ca = [ca];
            console.log(
              "📜 Using CA certificate:",
              this.config.connection.caPath
            );
          } else {
            console.log(
              "📜 CA certificate path specified but file not found, proceeding without it"
            );
          }
        } catch (err) {
          console.warn(
            "⚠️  Warning: Could not load CA certificate:",
            err.message
          );
        }
      }

      console.log("🔗 Attempting TLS connection...");
      this.client = tls.connect(options, () => {
        console.log("✅ TLS connection established");
        console.log("🔒 Connection details:");
        console.log("   Authorized:", this.client.authorized);
        console.log("   Cipher:", this.client.getCipher().name);
        console.log("   Protocol:", this.client.getProtocol());
        console.log(
          "   Server Certificate Subject:",
          this.client.getPeerCertificate().subject
        );

        if (!this.client.authorized) {
          console.log(
            "⚠️  Certificate not authorized:",
            this.client.authorizationError
          );
          if (this.config.connection.rejectUnauthorized !== false) {
            reject(
              new Error(
                "TLS certificate not authorized: " +
                  this.client.authorizationError
              )
            );
            return;
          }
        }

        resolve();
      });

      this.client.on("error", (error) => {
        console.error("❌ TLS connection error:", error.message);
        if (error.code === "DEPTH_ZERO_SELF_SIGNED_CERT") {
          console.log(
            "💡 Tip: Set 'rejectUnauthorized: false' in config for self-signed certificates"
          );
        }
        reject(error);
      });

      this.client.on("close", () => {
        console.log("🔌 TLS connection closed");
        this.authenticated = false;
      });

      this.client.on("secureConnect", () => {
        console.log("🔐 Secure connection established");
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

      console.log("🔑 Sending authentication (encrypted)...");

      const responseHandler = (data) => {
        const response = data.toString().trim();
        console.log("✅ Auth response received:", response);
        this.client.removeListener("data", responseHandler);

        try {
          const parsed = JSON.parse(response);
          if (parsed.error) {
            reject(new Error(parsed.error));
          } else {
            this.authenticated = true;
            console.log("🎉 Authentication successful - connection secured!");
            resolve();
          }
        } catch (e) {
          console.error("❌ Failed to parse auth response:", e.message);
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
          .replace("{{testNumber}}", (i + 1).toString())
          .replace("{{connectionType}}", "🔒 TLS Encrypted"),
      };

      console.log(`📧 Sending encrypted email ${i + 1}/${count}:`, {
        to: emailData.to,
        subject: emailData.subject,
      });

      try {
        await this.sendSingleEmail(emailData);
        results.success++;
        if (count > 1) {
          console.log(
            `✅ Email ${i + 1}/${count} sent successfully (encrypted)`
          );
        }

        if (i < count - 1 && this.config.settings.defaultDelay) {
          console.log(
            `⏳ Waiting ${this.config.settings.defaultDelay}ms before next email...`
          );
          await new Promise((resolve) =>
            setTimeout(resolve, this.config.settings.defaultDelay)
          );
        }
      } catch (error) {
        results.failed++;
        console.error(`❌ Failed to send email ${i + 1}:`, error.message);
      }
    }

    return results;
  }

  sendSingleEmail(emailData) {
    return new Promise((resolve, reject) => {
      const data = JSON.stringify(emailData) + "\n";
      console.log("📤 Sending encrypted email data...");

      const responseHandler = (data) => {
        const response = data.toString().trim();
        console.log("📥 Email response received:", response);
        this.client.removeListener("data", responseHandler);

        try {
          const parsed = JSON.parse(response);
          if (parsed.error) {
            reject(new Error(parsed.error));
          } else {
            resolve(parsed.message || "Success");
          }
        } catch (e) {
          console.error("❌ Failed to parse email response:", e.message);
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
      console.log("🔌 Disconnected from TLS server");
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
    console.log("\n🔒 GoMailer TLS Test Client");
    console.log("=============================");

    while (true) {
      console.log("\n📋 Options:");
      console.log("1. 📧 Send single encrypted email");
      console.log("2. 📬 Send multiple encrypted emails");
      console.log("3. 🔍 Show TLS connection info");
      console.log("4. ❌ Exit");

      const answer = await this.question("\n👉 Choose an option (1-4): ");

      if (answer === "4") {
        await this.disconnect();
        break;
      }

      try {
        if (!this.client) {
          await this.connect();
        }

        if (answer === "1") {
          const result = await this.sendEmail(1);
          console.log("\n📊 Result:", result);
        } else if (answer === "2") {
          const countStr = await this.question(
            "📬 How many emails do you want to send? "
          );
          const count = parseInt(countStr);

          if (isNaN(count) || count <= 0) {
            console.log("❌ Please enter a valid number greater than 0");
            continue;
          }

          const result = await this.sendEmail(count);
          console.log("\n📊 Final Results:");
          console.log(`✅ Successful: ${result.success}`);
          console.log(`❌ Failed: ${result.failed}`);
        } else if (answer === "3") {
          if (this.client && this.client.authorized !== undefined) {
            console.log("\n🔒 TLS Connection Information:");
            console.log("================================");
            console.log(
              "Status:",
              this.client.authorized ? "✅ Authorized" : "⚠️  Not Authorized"
            );
            console.log("Cipher:", this.client.getCipher().name);
            console.log("Protocol:", this.client.getProtocol());
            const cert = this.client.getPeerCertificate();
            console.log("Server Certificate:");
            console.log("  Subject:", cert.subject);
            console.log("  Issuer:", cert.issuer);
            console.log("  Valid From:", cert.valid_from);
            console.log("  Valid To:", cert.valid_to);
          } else {
            console.log("❌ No active TLS connection");
          }
        } else {
          console.log("❌ Invalid option. Please choose 1, 2, 3, or 4.");
        }
      } catch (error) {
        console.error("❌ Error:", error.message);
        await this.disconnect();
      }
    }
  }
}

async function main() {
  const client = new EmailTestTLSClient();

  try {
    await client.initialize();
    await client.showMenu();
  } catch (error) {
    console.error("❌ Fatal error:", error.message);
    process.exit(1);
  }
}

// Handle graceful shutdown
process.on("SIGINT", () => {
  console.log("\n👋 Shutting down gracefully...");
  process.exit(0);
});

if (require.main === module) {
  main();
}

module.exports = EmailTestTLSClient;
