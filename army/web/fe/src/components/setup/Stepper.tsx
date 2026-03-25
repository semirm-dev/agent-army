import { Check } from 'lucide-react';
import { cn } from '@/lib/utils';

interface StepperProps {
  steps: string[];
  currentStep: number;
  onStepClick: (step: number) => void;
}

export function Stepper({ steps, currentStep, onStepClick }: StepperProps) {
  return (
    <div className="flex items-center gap-2">
      {steps.map((label, i) => {
        const isCompleted = i < currentStep;
        const isCurrent = i === currentStep;
        return (
          <div key={label} className="flex items-center gap-2">
            {i > 0 && (
              <div
                className={cn(
                  'h-px w-6',
                  i <= currentStep ? 'bg-primary' : 'bg-border'
                )}
              />
            )}
            <button
              onClick={() => isCompleted && onStepClick(i)}
              disabled={!isCompleted}
              className={cn(
                'flex items-center gap-1.5 text-xs font-medium transition-colors',
                isCompleted && 'cursor-pointer text-primary hover:text-primary/80',
                isCurrent && 'text-primary',
                !isCompleted && !isCurrent && 'text-muted-foreground cursor-default'
              )}
            >
              <span
                className={cn(
                  'size-6 rounded-full flex items-center justify-center text-[11px] font-bold shrink-0',
                  isCompleted && 'bg-primary text-primary-foreground',
                  isCurrent && 'bg-primary text-primary-foreground',
                  !isCompleted && !isCurrent && 'bg-muted text-muted-foreground'
                )}
              >
                {isCompleted ? <Check className="size-3" /> : i + 1}
              </span>
              <span className="hidden sm:inline">{label}</span>
            </button>
          </div>
        );
      })}
    </div>
  );
}
