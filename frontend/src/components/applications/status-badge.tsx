import { Badge } from "@/components/ui/badge";
import {
  APPLICATION_PIPELINE_STATUS_OPTIONS,
  type ApplicationStatus,
} from "@/types/application";

const statusVariants: Record<
  ApplicationStatus,
  "default" | "secondary" | "destructive" | "outline"
> = {
  draft: "secondary",
  applied: "default",
  replied: "outline",
  interview: "outline",
  offer: "default",
  rejected: "destructive",
  ghosted: "secondary",
};

const labels = APPLICATION_PIPELINE_STATUS_OPTIONS.reduce(
  (acc, { value, label }) => {
    acc[value] = label;
    return acc;
  },
  {} as Record<ApplicationStatus, string>
);

export function StatusBadge({ status }: { status: ApplicationStatus }) {
  const variant = statusVariants[status] ?? statusVariants.draft;
  const label = labels[status] ?? labels.draft;
  return <Badge variant={variant}>{label}</Badge>;
}
