export interface SyncAction {
  type: 'install' | 'remove';
  item_type: 'plugin' | 'skill';
  name: string;
  source: string;
  destination: string;
}

export type SyncEvent =
  | { event: 'plan'; actions: SyncAction[] }
  | { event: 'action_start'; type: string; item_type: string; name: string }
  | {
      event: 'action_done';
      type: string;
      item_type: string;
      name: string;
      success: boolean;
      error?: string;
    }
  | { event: 'complete'; succeeded: number; failed: number }
  | { event: 'exit'; code: number }
  | { event: 'error'; message: string };
