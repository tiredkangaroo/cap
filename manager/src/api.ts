import { Config } from "./types.ts";

export class Proxy {
    loaded: boolean = false;
    name: string;
    url: string;
    config: Config;
    // token for accessing the proxy will be set here as well

    constructor(name: string, url: string) {
        this.name = name;
        this.url = url;
        this.config = {
            real_ip_header: false,
            certificate_lifetime: 0,
            mitm: false,
        };
    }

    async Init(): Promise<void> {
        this.config = await this.getConfig();
        this.loaded = true;
    }

    async getConfig(): Promise<Config> {
        const response = await fetch(`${this.url}/config`);
        if (!response.ok) {
            throw new Error(`failed to fetch config: ${response.statusText}`);
        }
        const config = await response.json();
        this.config = config;
        this.loaded = true;
        return config;
    }

    async setConfig(config: Config): Promise<void> {
        const response = await fetch(`${this.url}/config`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(config),
        });
        if (!response.ok) {
            throw new Error(`failed to set config: ${response.statusText}`);
        }
    }
}
