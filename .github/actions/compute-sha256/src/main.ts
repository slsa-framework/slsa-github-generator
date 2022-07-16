import * as core from '@actions/core'
import * as fs from 'fs'
import * as crypto from 'crypto'

function shasum256(untrustedPath: string): string {
  if (!fs.existsSync(untrustedPath)) {
    throw new Error(`File ${untrustedPath} not present`)
  }
  const untrustedFile = fs.readFileSync(untrustedPath)
  return crypto.createHash('sha256').update(untrustedFile).digest('hex')
}

async function run(): Promise<void> {
  // Get the path to the untrusted file from ENV variable 'UNTRUSTED_PATH'
  const untrustedPath = core.getInput('untrusted_path')
  const sha = shasum256(untrustedPath)
  core.setOutput('sha256', sha)
}
run()
