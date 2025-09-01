import React, { useState, useEffect } from 'react';

import { api, Category, Sound } from '../api';
import Button from './button';

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
        <li className={active ? 'active' : ''}>
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
        </li>
    );
}

export default SoundAdminOption;
