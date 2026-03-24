import { Plus, Check } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

interface CatalogCardProps {
  name: string;
  description: string;
  source: string;
  tags: string[];
  inManifest: boolean;
  onAdd: () => void;
  isAdding: boolean;
}

export function CatalogCard({
  name,
  description,
  source,
  tags,
  inManifest,
  onAdd,
  isAdding,
}: CatalogCardProps) {
  return (
    <Card className="flex flex-col">
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{name}</CardTitle>
        <CardDescription className="line-clamp-2">{description}</CardDescription>
      </CardHeader>
      <CardContent className="flex-1">
        <p className="text-xs text-muted-foreground mb-2 truncate" title={source}>
          {source}
        </p>
        <div className="flex flex-wrap gap-1">
          {tags.map((tag) => (
            <Badge key={tag} variant="secondary" className="text-[10px]">
              {tag}
            </Badge>
          ))}
        </div>
      </CardContent>
      <CardFooter>
        {inManifest ? (
          <Button variant="ghost" size="sm" disabled className="w-full">
            <Check className="size-4" />
            In Manifest
          </Button>
        ) : (
          <Button
            variant="outline"
            size="sm"
            className="w-full"
            onClick={onAdd}
            disabled={isAdding}
          >
            <Plus className="size-4" />
            {isAdding ? 'Adding...' : 'Add to Manifest'}
          </Button>
        )}
      </CardFooter>
    </Card>
  );
}
