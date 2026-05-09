/**
 * Output Panel Formatter
 * Displays tool responses with syntax highlighting, status colors, timestamps
 */

export interface FormattedOutput {
  html?: string;
  text: string;
  status: 'success' | 'error' | 'warning' | 'info';
  timestamp: string;
}

export class OutputFormatter {
  /**
   * Format successful tool response
   */
  static formatSuccess(
    output: any,
    toolName: string,
    durationMs: number,
    timestamp: string
  ): FormattedOutput {
    let text = `✅ ${toolName} executed successfully\n\n`;
    text += `**Duration**: ${durationMs}ms\n`;
    text += `**Timestamp**: ${new Date(timestamp).toLocaleString()}\n\n`;

    // Format output based on tool
    if (toolName === 'place-order') {
      text += this.formatOrderResponse(output);
    } else if (toolName === 'close-position') {
      text += this.formatClosePositionResponse(output);
    } else if (toolName === 'positions-list') {
      text += this.formatPositionsResponse(output);
    } else {
      text += `### Output\n\`\`\`json\n${JSON.stringify(output, null, 2)}\n\`\`\``;
    }

    return {
      text,
      status: 'success',
      timestamp,
    };
  }

  /**
   * Format error response
   */
  static formatError(
    error: any,
    toolName: string,
    durationMs: number,
    timestamp: string
  ): FormattedOutput {
    let text = `❌ ${toolName} failed\n\n`;
    text += `**Error**: ${error.code || error.message}\n`;
    text += `**Message**: ${error.message}\n`;
    text += `**Duration**: ${durationMs}ms\n`;
    text += `**Timestamp**: ${new Date(timestamp).toLocaleString()}\n\n`;

    if (error.details) {
      text += `### Details\n\`\`\`json\n${JSON.stringify(error.details, null, 2)}\n\`\`\``;
    }

    text += `\n\n[Retry] [View Logs]`;

    return {
      text,
      status: 'error',
      timestamp,
    };
  }

  /**
   * Format warning response
   */
  static formatWarning(message: string, timestamp: string): FormattedOutput {
    const text = `⚠️ Warning\n\n${message}\n\n[Acknowledge]`;

    return {
      text,
      status: 'warning',
      timestamp,
    };
  }

  /**
   * Format info message
   */
  static formatInfo(message: string, timestamp: string): FormattedOutput {
    const text = `ℹ️ ${message}`;

    return {
      text,
      status: 'info',
      timestamp,
    };
  }

  /**
   * Format place-order response
   */
  private static formatOrderResponse(output: any): string {
    let text = `### Order Filled\n\n`;
    text += `| Field | Value |\n`;
    text += `|-------|-------|\n`;
    text += `| **Ticket** | ${output.ticket} |\n`;
    text += `| **Status** | ${output.status} |\n`;
    text += `| **Entry Price** | ${output.entryPrice} |\n`;
    text += `| **Commission** | ${output.commission || 'N/A'} |\n`;
    text += `| **Swap** | ${output.swap || 'N/A'} |\n`;

    return text;
  }

  /**
   * Format close-position response
   */
  private static formatClosePositionResponse(output: any): string {
    let text = `### Position Closed\n\n`;
    text += `| Field | Value |\n`;
    text += `|-------|-------|\n`;
    text += `| **Ticket** | ${output.ticket} |\n`;
    text += `| **Status** | ${output.status} |\n`;
    text += `| **Close Price** | ${output.closePrice} |\n`;
    text += `| **P&L** | ${output.pnl >= 0 ? '✅' : '❌'} $${output.pnl} |\n`;
    text += `| **Commission** | ${output.commission || 'N/A'} |\n`;

    return text;
  }

  /**
   * Format positions-list response
   */
  private static formatPositionsResponse(positions: any[]): string {
    if (!Array.isArray(positions) || positions.length === 0) {
      return `No open positions.`;
    }

    let text = `### Open Positions (${positions.length})\n\n`;
    text += `| Ticket | Symbol | Type | Volume | Entry | Current | P&L | P&L % |\n`;
    text += `|--------|--------|------|--------|-------|---------|-----|-------|\n`;

    positions.forEach((pos) => {
      const pnlColor = pos.pnl >= 0 ? '✅' : '❌';
      text += `| ${pos.ticket} | ${pos.symbol} | ${pos.type} | ${pos.volume} | ${pos.entryPrice} | ${pos.currentPrice} | ${pnlColor} $${pos.pnl} | ${pos.pnlPercent}% |\n`;
    });

    return text;
  }

  /**
   * Format as HTML for rich display
   */
  static toHtml(formatted: FormattedOutput): string {
    const statusColor = {
      success: '#4ec9b0',
      error: '#d4534f',
      warning: '#f4d03f',
      info: '#569cd6',
    }[formatted.status];

    let html = `<div style="border-left: 4px solid ${statusColor}; padding: 12px; background: #f5f5f5;">`;
    html += `<div style="color: ${statusColor}; font-weight: bold; margin-bottom: 8px;">`;
    html += `${formatted.text.split('\n')[0]}`;
    html += `</div>`;
    html += `<div style="color: #666; white-space: pre-wrap; font-family: monospace; font-size: 12px;">`;
    html += formatted.text.replace(/\n/g, '<br/>');
    html += `</div>`;
    html += `</div>`;

    return html;
  }

  /**
   * Format execution summary for panel
   */
  static formatPanel(
    outputs: FormattedOutput[]
  ): string {
    if (outputs.length === 0) {
      return 'No executions yet.';
    }

    let panel = `## Execution History\n\n`;

    outputs.forEach((output, idx) => {
      panel += `### ${idx + 1}. ${output.status.toUpperCase()}\n`;
      panel += `${output.text}\n\n`;
      panel += `---\n\n`;
    });

    return panel;
  }
}
