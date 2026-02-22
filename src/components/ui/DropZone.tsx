import { useRef } from 'react';
import './DropZone.css';

interface DropZoneProps {
    file: File | null;
    disabled?: boolean;
    onChange: (file: File) => void;
}

export function DropZone({ file, disabled = false, onChange }: DropZoneProps): React.ReactElement {
    const inputRef = useRef<HTMLInputElement>(null);

    function handleChange(e: React.ChangeEvent<HTMLInputElement>): void {
        const picked = e.target.files?.[0];
        if (picked) onChange(picked);
    }

    return (
        <div
            className={`dz-zone${file ? ' dz-zone--ready' : ''}`}
            onClick={() => !disabled && inputRef.current?.click()}
        >
            <input
                ref={inputRef}
                type="file"
                accept="image/*"
                className="dz-input"
                onChange={handleChange}
                disabled={disabled}
            />

            {disabled ? (
                <div className="dz-spinner" />
            ) : file ? (
                <div className="dz-file">
                    <div className="dz-file-name">{file.name}</div>
                    <div className="dz-file-meta">
                        {(file.size / 1024 / 1024).toFixed(1)} MB · click to change
                    </div>
                </div>
            ) : (
                <div className="dz-prompt">
                    <div className="dz-prompt-arrow">↑</div>
                    <div className="dz-prompt-text">Click to select image</div>
                </div>
            )}
        </div>
    );
}
