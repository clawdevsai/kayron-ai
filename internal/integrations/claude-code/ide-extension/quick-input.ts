/**
 * Quick Input Dialog for Order Parameters
 * Multi-step form: symbol → volume → type → price
 */

export interface QuickInputState {
  symbol?: string;
  volume?: number;
  type?: 'BUY' | 'SELL';
  price?: string;
}

export interface QuickInputStep {
  title: string;
  prompt: string;
  placeholder: string;
  validate?: (value: string) => { valid: boolean; error?: string };
  description?: string;
  type?: 'input' | 'dropdown' | 'number';
  options?: Array<{ label: string; value: string }>;
}

export class QuickInputDialog {
  private state: QuickInputState = {};
  private currentStep: number = 0;
  private steps: QuickInputStep[] = [];

  constructor() {
    this.initializeSteps();
  }

  /**
   * Initialize multi-step form
   */
  private initializeSteps(): void {
    this.steps = [
      {
        title: 'Trading Symbol',
        prompt: 'Enter trading pair (e.g., EURUSD, GBPUSD)',
        placeholder: 'EURUSD',
        description: 'The instrument to trade',
        type: 'input',
        validate: (value) => {
          if (!value || value.length === 0) {
            return { valid: false, error: 'Symbol is required' };
          }
          if (!/^[A-Z]{6}$/.test(value)) {
            return { valid: false, error: 'Symbol must be 6 letters (e.g., EURUSD)' };
          }
          return { valid: true };
        },
      },
      {
        title: 'Volume',
        prompt: 'Enter lot size (0.01 - 100)',
        placeholder: '0.1',
        description: 'Trading volume in lots',
        type: 'number',
        validate: (value) => {
          const num = parseFloat(value);
          if (isNaN(num)) {
            return { valid: false, error: 'Volume must be a number' };
          }
          if (num < 0.01 || num > 100) {
            return { valid: false, error: 'Volume must be between 0.01 and 100' };
          }
          return { valid: true };
        },
      },
      {
        title: 'Order Type',
        prompt: 'Choose order direction',
        placeholder: 'Select BUY or SELL',
        description: 'Buy to open long, Sell to open short',
        type: 'dropdown',
        options: [
          { label: 'BUY - Open Long', value: 'BUY' },
          { label: 'SELL - Open Short', value: 'SELL' },
        ],
      },
      {
        title: 'Price',
        prompt: 'Enter price or MARKET (optional)',
        placeholder: 'MARKET',
        description: 'MARKET for instant fill, or specific price for limit order',
        type: 'input',
        validate: (value) => {
          if (!value) {
            return { valid: true }; // Optional
          }
          if (value.toUpperCase() === 'MARKET') {
            return { valid: true };
          }
          const num = parseFloat(value);
          if (isNaN(num)) {
            return { valid: false, error: 'Price must be a number or "MARKET"' };
          }
          if (num <= 0) {
            return { valid: false, error: 'Price must be positive' };
          }
          return { valid: true };
        },
      },
    ];
  }

  /**
   * Get current step
   */
  getCurrentStep(): QuickInputStep {
    return this.steps[this.currentStep];
  }

  /**
   * Process input for current step
   */
  processInput(value: string): { valid: boolean; error?: string } {
    const step = this.getCurrentStep();

    if (step.validate) {
      const validation = step.validate(value);
      if (!validation.valid) {
        return validation;
      }
    }

    // Store value
    const stepNames: (keyof QuickInputState)[] = ['symbol', 'volume', 'type', 'price'];
    const key = stepNames[this.currentStep];

    if (key === 'volume') {
      this.state[key] = parseFloat(value);
    } else {
      (this.state as any)[key] = value;
    }

    return { valid: true };
  }

  /**
   * Move to next step
   */
  nextStep(): boolean {
    if (this.currentStep < this.steps.length - 1) {
      this.currentStep++;
      return true;
    }
    return false;
  }

  /**
   * Move to previous step
   */
  previousStep(): boolean {
    if (this.currentStep > 0) {
      this.currentStep--;
      return true;
    }
    return false;
  }

  /**
   * Check if all required steps completed
   */
  isComplete(): boolean {
    return (
      this.state.symbol &&
      this.state.volume !== undefined &&
      this.state.type &&
      this.currentStep === this.steps.length - 1
    );
  }

  /**
   * Get collected parameters
   */
  getParameters(): QuickInputState {
    return { ...this.state };
  }

  /**
   * Show account info context
   */
  getAccountContext(): string {
    return (
      `Account Balance: $10,000\n` +
      `Available Margin: $8,500\n` +
      `Used Margin: $1,500\n` +
      `Margin Level: 566%\n\n` +
      `Current Positions: 2\n` +
      `Open Trades: EURUSD, GBPUSD`
    );
  }

  /**
   * Format dialog display
   */
  formatDialog(): string {
    const step = this.getCurrentStep();
    const progress = `Step ${this.currentStep + 1} of ${this.steps.length}`;

    let dialog = `## Place Order\n\n`;
    dialog += `${progress}\n\n`;
    dialog += `### ${step.title}\n`;
    dialog += `${step.prompt}\n\n`;

    if (step.type === 'dropdown' && step.options) {
      dialog += `Options:\n`;
      step.options.forEach((opt) => {
        dialog += `- ${opt.label} (\`${opt.value}\`)\n`;
      });
    } else {
      dialog += `Placeholder: \`${step.placeholder}\`\n`;
    }

    if (step.description) {
      dialog += `\n_${step.description}_\n`;
    }

    // Show account context on volume step
    if (this.currentStep === 1) {
      dialog += `\n### Account Info\n`;
      dialog += this.getAccountContext();
    }

    return dialog;
  }

  /**
   * Format summary before execution
   */
  formatSummary(): string {
    let summary = `## Order Summary\n\n`;
    summary += `**Symbol**: ${this.state.symbol}\n`;
    summary += `**Volume**: ${this.state.volume} lots\n`;
    summary += `**Type**: ${this.state.type}\n`;
    summary += `**Price**: ${this.state.price || 'MARKET'}\n\n`;
    summary += `[Execute] [Back] [Cancel]`;

    return summary;
  }

  /**
   * Reset dialog
   */
  reset(): void {
    this.state = {};
    this.currentStep = 0;
  }
}
