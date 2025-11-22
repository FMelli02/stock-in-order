import React from 'react';

interface PaginationMetadata {
  current_page: number;
  page_size: number;
  first_page: number;
  last_page: number;
  total_records: number;
}

interface PaginationControlsProps {
  metadata: PaginationMetadata;
  onPageChange: (page: number) => void;
}

const PaginationControls: React.FC<PaginationControlsProps> = ({ metadata, onPageChange }) => {
  const { current_page, last_page, total_records } = metadata;

  const handlePrevious = () => {
    if (current_page > 1) {
      onPageChange(current_page - 1);
    }
  };

  const handleNext = () => {
    if (current_page < last_page) {
      onPageChange(current_page + 1);
    }
  };

  const handleFirst = () => {
    if (current_page > 1) {
      onPageChange(1);
    }
  };

  const handleLast = () => {
    if (current_page < last_page) {
      onPageChange(last_page);
    }
  };

  // No mostrar controles si no hay registros
  if (total_records === 0) {
    return (
      <div className="flex items-center justify-center py-4 text-gray-500">
        No hay registros para mostrar
      </div>
    );
  }

  return (
    <div className="flex items-center justify-between py-4 px-4 border-t border-gray-200">
      <div className="text-sm text-gray-700">
        Mostrando página <span className="font-medium">{current_page}</span> de{' '}
        <span className="font-medium">{last_page}</span> · Total:{' '}
        <span className="font-medium">{total_records}</span> registros
      </div>

      <div className="flex items-center space-x-2">
        <button
          onClick={handleFirst}
          disabled={current_page === 1}
          className="px-3 py-1 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Primera
        </button>

        <button
          onClick={handlePrevious}
          disabled={current_page === 1}
          className="px-3 py-1 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Anterior
        </button>

        <span className="text-sm text-gray-700 px-2">
          Página {current_page} de {last_page}
        </span>

        <button
          onClick={handleNext}
          disabled={current_page === last_page}
          className="px-3 py-1 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Siguiente
        </button>

        <button
          onClick={handleLast}
          disabled={current_page === last_page}
          className="px-3 py-1 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Última
        </button>
      </div>
    </div>
  );
};

export default PaginationControls;
