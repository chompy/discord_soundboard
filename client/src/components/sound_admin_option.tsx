import React from 'react';
import { SortableKnob } from 'react-easy-sort';

export type SoundAdminOptionProperties = {
    label: string;
    active?: boolean;
    onClick?: () => void;
    onEdit?: () => void;
    onDelete?: () => void;
};

function SoundAdminOption({
    label,
    active,
    onClick,
    onEdit,
    onDelete,
}: SoundAdminOptionProperties) {
    return (
        <div className={`sound-admin-option ${active && 'active'}`}>
            <SortableKnob>
                <div className="handle"></div>
            </SortableKnob>
            <span
                title={label}
                onClick={(e) => {
                    e.preventDefault();
                    onClick?.();
                }}
            >
                {label}
            </span>
            <a
                href="#"
                onClick={(e) => {
                    e.preventDefault();
                    onDelete?.();
                }}
            >
                ❌
            </a>
            <a
                href="#"
                onClick={(e) => {
                    e.preventDefault();
                    onEdit?.();
                }}
            >
                ✏️
            </a>
        </div>
    );
}

export default SoundAdminOption;
