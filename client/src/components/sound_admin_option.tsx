import React from 'react';
import { SortableKnob } from 'react-easy-sort';

export type SoundAdminOptionProperties = {
    label: string;
    active?: boolean;
    onClick?: () => void;
    onEdit?: () => void;
    onDelete?: () => void;
    onDownload?: () => void;
};

function SoundAdminOption({
    label,
    active,
    onClick,
    onEdit,
    onDelete,
    onDownload,
}: SoundAdminOptionProperties) {
    return (
        <div className={`sound-admin-option${active ? ' active' : ''}`}>
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
                title='Delete'
                onClick={(e) => {
                    e.preventDefault();
                    if (confirm(`Are you sure you want to delete '${label}'?`)) {
                        onDelete?.();
                    }
                }}
            >
                ❌
            </a>
            <a
                href="#"
                title='Rename'
                onClick={(e) => {
                    e.preventDefault();
                    onEdit?.();
                }}
            >
                ✏️
            </a>
            {onDownload && 
            <a
                href="#"
                title='Download'
                onClick={(e) => {
                    e.preventDefault();
                    onDownload?.();
                }}
            >
                💾
            </a>}
            

        </div>
    );
}

export default SoundAdminOption;
