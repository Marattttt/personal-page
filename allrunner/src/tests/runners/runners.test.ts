import pino from "pino"
import { LangNotSupportedError, Runner, RunResult } from "../../runners/runner"

jest.mock('pino')
const logger = pino()

const helloWorld: RunResult = {
	stdout: 'Hello world',
	stderr: '',
	exitCode: 0,
	execTimeMs: 1
} as const

const jsMock = { runjs: jest.fn(async (_) => helloWorld) }
const goMock = { rungo: jest.fn(async (_) => helloWorld) }

test('Runner forwards to js correctly', async () => {
	const runner = new Runner(jsMock)

	expect(await runner.run(logger, 'js', '', 1000))
		.toBe<RunResult>(helloWorld)
})

test('Runner forwards to go correctly', async () => {
	const runner = new Runner(undefined, goMock)

	expect(await runner.run(logger, 'go', '', 1000))
		.toBe<RunResult>(helloWorld)
})

test('Runner throws on unexpected language', async () => {
	const runner = new Runner(jsMock)

	const langname = 'some other language'
	expect(runner.run(logger, langname as any, '', 1000))
		.rejects
		.toThrow(new LangNotSupportedError(langname))
})
