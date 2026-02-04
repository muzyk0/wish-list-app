import Image from "next/image";

interface GiftItemImageProps {
  src: string;
  alt: string;
  className?: string;
}

export default function GiftItemImage({
  src,
  alt,
  className,
}: GiftItemImageProps) {
  if (!src) {
    return (
      <div className={`bg-muted flex items-center justify-center ${className}`}>
        <div className="text-muted-foreground">No image</div>
      </div>
    );
  }

  return (
    <div className={className}>
      <Image
        src={src}
        alt={alt}
        fill
        className="object-cover rounded-md"
        sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
      />
    </div>
  );
}
