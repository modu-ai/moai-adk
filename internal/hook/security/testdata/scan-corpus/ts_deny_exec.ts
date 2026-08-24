import cp from 'child_process';

export function run(userInput: string): void {
  cp.exec(userInput);
}
