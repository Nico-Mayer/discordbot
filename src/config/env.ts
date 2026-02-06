import consola from "consola"

class Config {
	TOKEN: string
	SERVER_ID: string
	APP_ID: string
	LAVALINK_HOST: string
	LAVALINK_PORT: number
	LAVALINK_PASSWORD: string

	constructor() {
		this.TOKEN = this.getRequiredEnv("TOKEN")
		this.SERVER_ID = this.getRequiredEnv("SERVER_ID")
		this.APP_ID = this.getRequiredEnv("APP_ID")
		this.LAVALINK_HOST = this.getRequiredEnv("LAVALINK_HOST")
		this.LAVALINK_PORT = this.getRequiredEnvAsInt("LAVALINK_PORT")
		this.LAVALINK_PASSWORD = this.getRequiredEnv("LAVALINK_PASSWORD")
	}

	private getRequiredEnv(key: string): string {
		const value = process.env[key]

		if (value === undefined || value.length === 0) {
			consola.error(`Required env variable ${key} is not set`)
			process.exit(1)
		}
		return value
	}
	private getRequiredEnvAsInt(key: string): number {
		const value = this.getRequiredEnv(key)
		let intValue: number

		try {
			intValue = parseInt(value, 10)
		} catch (error) {
			consola.error(`Required ENV variable ${key} cant be parsed to int`, error)
			process.exit(1)
		}

		return intValue
	}
}

export default new Config()
